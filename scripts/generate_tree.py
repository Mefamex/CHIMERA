""" scripts/generate_tree.py
Generate a visual tree representation of the directory structure.
Saves the output to 'file-tree.md'.
"""

import os
from pathlib import Path
from rich import print
from rich.markup import escape
from rich.text import Text
from rich.tree import Tree
from rich.console import Console

def walk_directory(directory: Path, tree: Tree):
    # Sıralama: Önce klasörler, sonra dosyalar (Alfabetik)
    paths = sorted(
        directory.iterdir(),
        key=lambda path: (path.is_file(), path.name.lower())
    )
    
    for path in paths:
        # --- GİZLENECEK KLASÖRLER (Ignore List) ---
        if path.name.startswith(".") or path.name in ["__pycache__", "venv", "node_modules", "dist", "bin", "obj"]: continue
        
        # Dosya/Klasör Görünümü
        if path.is_dir():
            style = "dim" if path.name.startswith("__") else ""
            branch = tree.add(
                f"[bold magenta]:open_file_folder: {escape(path.name)}[/bold magenta]",
                style=style,
                guide_style=style,
            )
            walk_directory(path, branch)
        else:
            text_filename = Text(path.name, "green")
            text_filename.highlight_regex(r"\..*$", "bold red")
            tree.add(text_filename)

# Kullanım
root_path = Path(".") # Şu anki dizin
tree = Tree(
    f":file_folder: [bold white]{root_path.resolve().name}[/bold white]",
    guide_style="bold bright_blue",
)

walk_directory(root_path, tree)
# Render and save tree
console = Console(record=True)
console.print(tree)

md_content = "```text\n" + console.export_text() + "\n```\n"
with open("file-tree.md", "w", encoding="utf-8") as f:
    f.write(md_content)
    print(f"Saved tree to {os.path.abspath(f.name)}")

Console().print(tree)

