"""scripts/generate_tree.py
Generate a plain ASCII tree into `file-tree.md`.
Rules implemented:
- Within each directory, list files first (alphabetical), then subdirectories.
- After finishing a directory, if its last child is a file, insert a single
    blank line aligned to the parent's depth (i.e. the prefix used for this
    directory's entries).
- Load simple ignore names from `.gitignore` (no globs, only plain names/paths).
"""

from pathlib import Path
from rich import print
from rich.markup import escape
from rich.text import Text
from rich.tree import Tree
from rich.console import Console

# Defaults to ignore
DEFAULT_IGNORES = {"__pycache__", "venv", "node_modules", "dist", "bin", "obj"}

def load_ignores_from_gitignore(root: Path):
    gitignore = root / ".gitignore"
    ignores = set()
    if not gitignore.exists():
        return ignores
    try:
        for raw in gitignore.read_text(encoding="utf-8").splitlines():
            line = raw.strip()
            if not line or line.startswith("#"): continue
            if line.endswith("/"): line = line[:-1]
            if any(ch in line for ch in "*?[") or "/" in line or "\\" in line: continue
            name = Path(line).name
            if name:  ignores.add(name)
    except Exception: return set()
    return ignores

IGNORE_NAMES = DEFAULT_IGNORES | load_ignores_from_gitignore(Path("."))

def walk_directory(directory: Path, tree: Tree):
    paths = sorted(
        directory.iterdir(),
        key=lambda path: (not path.is_file(), path.name.lower())
    )
    
    for path in paths:
        if path.name.startswith(".") or path.name in IGNORE_NAMES: continue
            
        # Dosya/Klasör Görünümü
        if path.is_dir():
            style = "dim" if path.name.startswith("__") else ""
            branch = tree.add(
                f"[bold magenta]:open_file_folder: {escape(path.name)}[/bold magenta]",
                style=style,
                guide_style=style,
            )
            walk_directory(path, branch)
        else: tree.add(Text(path.name))

# Kullanım
root_path = Path(".") # Şu anki dizin
tree = Tree(
    f":file_folder: [bold white]{root_path.resolve().name}[/bold white]",
    guide_style="bold bright_blue",
)

walk_directory(root_path, tree)
print(tree)

output_path = root_path / "file-tree.md"
with output_path.open("w", encoding="utf-8") as f:
    f.write("```text\n\n")
    console = Console(file=f, force_terminal=False, color_system=None)
    console.print(tree)
    f.write("\n```\n")

print(f"\n[bold green]Dosya ağacı '{output_path.absolute()}' dosyasına kaydedildi.[/bold green]")