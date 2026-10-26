import hashlib
import json
from pathlib import Path
import tempfile
import unittest

from package_managers import generate


class PackageManagersTest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.dist = self.root / "dist"
        self.dist.mkdir()
        self.output = self.root / "packages"
        self.checksums = {}
        for artifact in ("source.tar.gz", "windows_amd64.zip", "windows_arm64.zip"):
            name = "eitri_1.2.3_" + artifact
            data = ("fixture:" + artifact).encode()
            (self.dist / name).write_bytes(data)
            self.checksums[name] = hashlib.sha256(data).hexdigest()
        self.write_checksums()

    def write_checksums(self):
        (self.dist / "checksums.txt").write_text(
            "".join(digest + "  " + name + "\n" for name, digest in self.checksums.items())
        )

    def test_packages_match_version_architecture_and_verified_hashes(self):
        generate("v1.2.3", self.dist, self.output)
        formula = (self.output / "Formula" / "eitri.rb").read_text()
        self.assertIn("/v1.2.3/eitri_1.2.3_source.tar.gz", formula)
        self.assertIn(self.checksums["eitri_1.2.3_source.tar.gz"], formula)
        self.assertIn('depends_on "go" => :build', formula)
        self.assertIn("main.version=#{version}", formula)
        scoop = json.loads((self.output / "bucket" / "eitri.json").read_text())
        self.assertEqual(scoop["version"], "1.2.3")
        self.assertEqual(scoop["bin"], "eitri.exe")
        for arch, go_arch in (("64bit", "amd64"), ("arm64", "arm64")):
            name = "eitri_1.2.3_windows_" + go_arch + ".zip"
            self.assertTrue(scoop["architecture"][arch]["url"].endswith("/v1.2.3/" + name))
            self.assertEqual(scoop["architecture"][arch]["hash"], self.checksums[name])

    def test_invalid_or_prerelease_tag_is_rejected(self):
        for tag in ("1.2.3", "v1.2.3-rc1", "v01.2.3", "v1.2.3/other"):
            with self.subTest(tag=tag), self.assertRaises(ValueError):
                generate(tag, self.dist, self.output)
        self.assertFalse(self.output.exists())

    def test_changed_artifact_is_rejected_before_writing_packages(self):
        (self.dist / "eitri_1.2.3_windows_arm64.zip").write_bytes(b"corrupt")
        with self.assertRaisesRegex(ValueError, "checksum mismatch"):
            generate("v1.2.3", self.dist, self.output)
        self.assertFalse(self.output.exists())

    def test_missing_checksum_is_rejected(self):
        del self.checksums["eitri_1.2.3_windows_amd64.zip"]
        self.write_checksums()
        with self.assertRaisesRegex(ValueError, "missing checksum"):
            generate("v1.2.3", self.dist, self.output)
        self.assertFalse(self.output.exists())

    def test_missing_artifact_is_rejected(self):
        (self.dist / "eitri_1.2.3_source.tar.gz").unlink()
        with self.assertRaises(FileNotFoundError):
            generate("v1.2.3", self.dist, self.output)
        self.assertFalse(self.output.exists())


if __name__ == "__main__":
    unittest.main()
