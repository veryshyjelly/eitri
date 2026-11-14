// Runtime contracts: config/config.go, config/validate.go, core/files.go,
// core/linker.go, core/judge.go, config/libraries.go and tmpl/funcs.go.
export const fields = [
 ['author','string','Optional; empty string','Attribution available to source/code templates. It has no effect unless your modifier uses .Author.','author: Your name'],
 ['editor','string template','Optional; empty disables automatic launch','A direct executable command rendered with the absolute .Filename. Quote the placeholder to preserve spaces. Runs from the project root; shell operators are not interpreted. Open files manually when empty.',`editor: 'code "{{ .Filename }}"'`],
 ['filename','mapping','Required','Contains URL-to-source routing rules.','filename:\n  rules: []'],
 ['filename.rules','list of mappings','Required; at least one rule','Rules are tried in order for the URL host until a regex matches. No matching rule produces an error.','filename:\n  rules:\n    - site: codeforces.com\n      regex: "problemset/problem/([0-9]+)/([A-Za-z0-9]+)"\n      template: "./src/{{ index .Captures 1 }}-{{ index .Captures 2 }}.cpp"'],
 ['filename.rules[].site','string','Required','Hostname, without a scheme or path. Matching ignores case and a leading www.; it is not a general URL pattern.','site: atcoder.jp'],
 ['filename.rules[].regex','Go regular expression','Required','Matches the full problem URL string. .Captures contains the entire match at index 0 and capture groups thereafter. Use YAML single quotes to avoid double-escaping backslashes.',"regex: 'problemset/task/(\u005cd+)'"],
 ['filename.rules[].template','string template','Required','Produces the editable solution path from .Captures and .Title. Relative output is resolved against the project root. Empty paths and filenames without extensions are rejected; creation makes parent directories. Existing files are preserved.','template: "./src/{{ index .Captures 1 }}.cpp"'],
 ['template','mapping','Optional unless creating files','Controls the initial editable file, separately from generated compiler input.','template:\n  source: ./base.cpp'],
 ['template.source','file path','Required when create_file is true; otherwise optional','Existing base source, resolved relative to the project root (absolute paths are allowed). A nonempty path is validated even if file creation is disabled.','source: ./base.cpp'],
 ['template.modifier','string template','Optional; {{ .Code }}','Renders the initial file from the base template. Available: .Author, .Time, .Url, .Title, .Filename and .Code. Here .Filename is the destination basename without its extension.','modifier: |\n  // {{ .Title }}\n  {{ .Code }}'],
 ['code','mapping','Optional','Controls the linked source used for tests and normal submission copying.','code:\n  modifier: "{{ .Code }}"'],
 ['code.modifier','string template','Optional; {{ .Code }}','Renders extracted solution code, ordered snippets and deduplicated headers. Available: .Author, .Time, .Url, .Code, .Header and .LibFiles. You must explicitly emit headers and snippets if you need them.','modifier: |\n  {{ .Header }}\n  {{ range .LibFiles }}\n  {{ .Content }}\n  {{ end }}\n  {{ .Code }}'],
 ['lib','mapping','Optional','Selects local snippet files and the expressions used to detect dependencies.','lib:\n  include: {}'],
 ['lib.regex','string template producing a Go regex','Required with nonempty lib.include','Rendered once per library with .Name; matches references in editable source and snippets. Snippet names are inserted literally by templates, so escape regex metacharacters in your naming/pattern conventions.','regex: "#include <{{ .Name }}.hpp>"'],
 ['lib.include','mapping of names to paths','Optional; empty map','A named file keeps its mapping key. A directory contributes immediate regular files under their basenames without extensions; nested directories are skipped. The directory mapping key is ignored. Every included path must exist. Duplicate library names are errors; keep non-source files out of snippet folders.', 'include:\n  dsu: ./lib/dsu.hpp\n  _: ./lib/extra/'],
 ['compiler','mapping','Required','Describes generated source, an optional build executable, argument list and sample run command.','compiler:\n  source: main.py\n  run: python3 main.py'],
 ['compiler.name','string','Optional; empty string','A descriptive label. It does not select a built-in compiler or language implementation.','name: My toolchain'],
 ['compiler.source','relative file path','Required','Where linked source is written inside a fresh temporary build workspace. May contain subdirectories; absolute or parent-escaping paths are rejected. This is distinct from template.source and the editable filename.','source: main.cpp'],
 ['compiler.compile','executable name or path','Optional; empty skips building','Executed directly with compiler.args from the temporary workspace, with a two-minute compilation deadline. Use a helper executable for project setup. This field is not a shell command line.','compile: g++'],
 ['compiler.args','list of strings','Optional; empty list','Arguments passed exactly as separate argv entries to compiler.compile. They are not Go templates or shell expansions. No arguments is valid for helpers with defaults.','args: ["-std=c++23", "-O2", "main.cpp", "-o", "main"]'],
 ['compiler.run','command string','Required','Split into an executable and arguments using shell-style quoting, then run directly in the temporary workspace for each sample. Pipes, redirects, variable expansion and chaining require an explicit shell/helper. Match the build output name.','run: ./main'],
 ['create_file','boolean','Optional; false','Create a missing editable file when a problem is selected. Existing files are never overwritten; requires template.source.','create_file: true'],
 ['run_on_save','boolean','Optional; false','Enable runs on source changes. The parent-directory watcher handles writes and atomic replacements, debounced by 200 ms. Manual r runs are still available.','run_on_save: true'],
 ['compiled_file','relative file path','Optional; empty uses linked source','For copying/submission, build first and read this text artifact from the temporary workspace instead of copying linked source. Must remain inside that workspace, including after symlink resolution, be a regular file, and fit within 4 MiB. Do not select a native executable as judge source.','compiled_file: generated/solution.cpp']
].map(([key,type,defaultValue,description,example])=>({key,type,defaultValue,description,example}));
export const contexts = [
 ['filename.rules[].template','.Captures (string list), .Title (string)','Title has single/double quotes removed. Captures[0] is the full regex match.'],
 ['template.modifier','.Author, .Time, .Url, .Title, .Filename, .Code (strings)','Code is the base file; Filename is the editable basename without its extension.'],
 ['code.modifier','.Author, .Time, .Url, .Code, .Header (strings), .LibFiles (list)','Each library has .Name and .Content. Code is the extracted editable source; Header contains unique header lines; libraries are dependency-first.'],
 ['lib.regex','.Name (string)','Rendered separately for each named snippet before regex compilation.'],
 ['editor','.Filename (string)','Absolute path to the editable file. No other template fields are provided.']
];
export const helpers = [
 ['toKebabCase','HelloWorld','hello-world'],['toCamelCase','hello_world','helloWorld'],
 ['toSnakeCase','HelloWorld','hello_world'],['toPascalCase','hello_world','HelloWorld'],
 ['toKEBABCase','helloWorld','HELLO-WORLD'],['toSNAKECase','helloWorld','HELLO_WORLD'],
 ['toLowerCase','Hello','hello'],['toUpperCase','Hello','HELLO'],
 ['stripPrefix','A. Watermelon','Watermelon'],['removeLines','first\nsecond','second']
];
