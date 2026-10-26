"""Generate Homebrew and Scoop definitions from verified GoReleaser artifacts."""

import argparse
import hashlib
import json
from pathlib import Path
import re
from string import Template


def validate_tag(tag):
    if not re.fullmatch(r"v(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)", tag):
        raise ValueError("package managers require a stable tag such as v1.0.1")


def generate(tag, dist, output):
    validate_tag(tag)
    version = tag[1:]
    dist, output = Path(dist), Path(output)
    checksums = {}
    for line in (dist / "checksums.txt").read_text().splitlines():
        digest, filename = line.split(maxsplit=1)
        filename = filename.lstrip(" *")
        if not re.fullmatch(r"[a-fA-F0-9]{64}", digest) or filename in checksums:
            raise ValueError("invalid or duplicate checksum for " + filename)
        checksums[filename] = digest.lower()

    def verified_hash(filename):
        expected = checksums.get(filename)
        if expected is None:
            raise ValueError("missing checksum for " + filename)
        with (dist / filename).open("rb") as artifact:
            digest = hashlib.sha256()
            for chunk in iter(lambda: artifact.read(1024 * 1024), b""):
                digest.update(chunk)
        if digest.hexdigest() != expected:
            raise ValueError("checksum mismatch for " + filename)
        return expected

    source_hash = verified_hash("eitri_{}_source.tar.gz".format(version))
    architecture = {}
    for scoop_arch, go_arch in (("64bit", "amd64"), ("arm64", "arm64")):
        filename = "eitri_{}_windows_{}.zip".format(version, go_arch)
        architecture[scoop_arch] = {
            "url": "https://github.com/veryshyjelly/eitri/releases/download/{}/{}".format(tag, filename),
            "hash": verified_hash(filename),
        }
    scoop = {
        "version": version,
        "description": "Competitive programming companion for the terminal",
        "homepage": "https://github.com/veryshyjelly/eitri",
        "license": "GPL-2.0-only",
        "architecture": architecture,
        "bin": "eitri.exe",
    }
    template = Path(__file__).resolve().parent.parent / "packaging" / "eitri.rb.tmpl"
    formula = Template(template.read_text()).substitute(version=version, sha256=source_hash)
    # Validate every input before writing either package definition.
    (output / "Formula").mkdir(parents=True, exist_ok=True)
    (output / "bucket").mkdir(parents=True, exist_ok=True)
    (output / "Formula" / "eitri.rb").write_text(formula)
    (output / "bucket" / "eitri.json").write_text(json.dumps(scoop, indent=2) + "\n")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--tag", required=True)
    parser.add_argument("--dist", type=Path, default=Path("dist"))
    parser.add_argument("--output", type=Path, default=Path("dist/package-managers"))
    args = parser.parse_args()
    try:
        generate(args.tag, args.dist, args.output)
    except (ValueError, OSError) as error:
        parser.exit(1, str(error) + "\n")
