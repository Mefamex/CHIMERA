import os, re
from pathlib import Path
from io import StringIO
import pathspec
from rich.tree import Tree
from rich.console import Console
from rich.markup import escape

def get_git_spec(root: Path):
    """.gitignore dosyasını okur ve pathspec nesnesi döner."""
    gitignore = root / ".gitignore"
    patterns = [".git"] # Varsayılan olarak .git'i yoksay
    if gitignore.exists(): patterns.extend(gitignore.read_text().splitlines())
    patterns.extend(["!file-tree.md"])
    return pathspec.PathSpec.from_lines('gitwildmatch', patterns)

def build_tree(directory: Path, tree: Tree, spec: pathspec.PathSpec, root_dir: Path):
    paths = sorted(directory.iterdir(), key=lambda p: (not p.is_file(), p.name.lower()))

    for path in paths:
        # Gitignore kontrolü (kök dizine göre göreceli yol ile)
        rel_path = path.relative_to(root_dir).as_posix()
        if spec.match_file(rel_path): continue
        if path.is_dir():
            branch = tree.add(f"[bold magenta]📂 {escape(path.name)}[/bold magenta]")
            build_tree(path, branch, spec, root_dir)
        else: tree.add(path.name)

# Ana İşlem
root_path = Path(".").resolve()
spec = get_git_spec(root_path)

tree = Tree(f"📁 [bold white]{root_path.name}[/bold white]", guide_style="bold bright_blue")
build_tree(root_path, tree, spec, root_path)

# Çıktıyı MD olarak kaydet
string_io = StringIO()
console = Console(file=string_io, width=100, color_system=None)
console.print(tree)

# Markdown dosyasına yaz (Regex kısmını sadeleştirdim)
output_content = "```text\n" + re.sub(r'(?m)^(┣━━ 📂)', lambda m: '┃\n' + m.group(1), string_io.getvalue()) + "```"
(root_path / "file-tree.md").write_text(output_content, encoding="utf-8")

print(f"Ağaç başarıyla oluşturuldu: file-tree.md")