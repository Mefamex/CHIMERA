#!/usr/bin/env python3
"""
Özelleştirilebilir dosya ağacı çıktısı.
Kullanım örnekleri README mesajında verilecek.
"""
import os
import argparse
import sys
from fnmatch import fnmatch

ROOT_DIR = "."
OUTPUT_FILE = "folder_tree.txt"
# Hariç tutma patternleri (virgülle ayrılmış, fnmatch destekli)
EXCLUDE_PATTERNS = []
# Yazılmasını istediğiniz .ile başlayan dosyalar (virgülle ayrılmış, isim bazlı)
INCLUDE_DOTFILES=[]
# Yalnızca ismi yazılacak .ile başlayan klasörler (virgülle ayrılmış, isim bazlı)
DOTDIRS_SHALLOW = [".venv", ".env", "node_modules" ,".git"]

def sorted_entries(path):
    try:
        names = os.listdir(path)
    except PermissionError:
        return []
    # sort: directories first, then files, both case-insensitive
    dirs = [n for n in names if os.path.isdir(os.path.join(path, n))]
    files = [n for n in names if not os.path.isdir(os.path.join(path, n))]
    return sorted(dirs, key=str.casefold) + sorted(files, key=str.casefold)


def should_exclude(name, exclude_patterns):
    if not exclude_patterns: return False
    for pat in exclude_patterns:
        if pat == name or fnmatch(name, pat): return True
    return False


def print_tree(root, exclude_patterns=None, include_dotfiles=None, dotdirs_shallow=None,indent = 2) -> list[str]:
    exclude_patterns = exclude_patterns or []
    include_dotfiles = set(include_dotfiles or [])
    dotdirs_shallow = set(dotdirs_shallow or [])
    output = []

    def _rec(path, prefix="", is_last=True, level=0):
        name = os.path.basename(path) if level>0 else os.path.basename(os.path.abspath(path)) or path
        connector = ("└" if is_last else "├") + "─"*(indent-1) +" "
        if level==0:
            print(f"{name}\n")
            output.append(name)
        else:
            display_name = name + ("/" if os.path.isdir(path) else "")
            print(f"{prefix}{connector}{display_name}\n")
            output.append(display_name)

        if not os.path.isdir(path):
            return

        baseprefix = prefix + (" " * indent if is_last else "│" + " " * (indent))
        try:
            entries = sorted_entries(path)
        except Exception:
            entries = []

        # filter entries according to rules
        filtered = []
        for e in entries:
            # global exclude by pattern/name
            if should_exclude(e, exclude_patterns):
                continue
            # dotfile/dotdir handling
            if e.startswith('.'):
                if os.path.isdir(os.path.join(path, e)):
                    # dot-directory
                    if e in dotdirs_shallow:
                        filtered.append(e)
                        continue
                    # else include and recurse only if not excluded
                    filtered.append(e)
                else:
                    # dot-file
                    if e in include_dotfiles:
                        filtered.append(e)
                    else:
                        continue
            else:
                filtered.append(e)

        for i, e in enumerate(filtered):
            full = os.path.join(path, e)
            last = (i == len(filtered)-1)
            # if it's a dotdir that should be shallow, print but don't recurse
            if e in dotdirs_shallow and os.path.isdir(full):
                connector2 = ("└" if last else "├") + "─"*(indent-1) +" "
                print(f"{baseprefix}{connector2}{e}/\n")
                output.append(f"{e}/")
                continue
            _rec(full, baseprefix, last, level+1)

    _rec(root, "", True, 0)
    return output


def parse_list_arg(s):
    if not s:
        return []
    return [p.strip() for p in s.split(',') if p.strip()]


def main(root_dir=None, output_file=None, exclude_patterns=None, include_dotfiles=None, dotdirs_shallow=None):    
    root_dir = root_dir or ROOT_DIR
    output_file = output_file or OUTPUT_FILE
    # Hariç tutma patternleri (virgülle ayrılmış, fnmatch destekli)
    exclude_patterns = exclude_patterns or EXCLUDE_PATTERNS
    # Yazılmasını istediğiniz .ile başlayan dosyalar (virgülle ayrılmış, isim bazlı)
    include_dotfiles = include_dotfiles or INCLUDE_DOTFILES
    # Yalnızca ismi yazılacak .ile başlayan klasörler (virgülle ayrılmış, isim bazlı)
    dotdirs_shallow = dotdirs_shallow or DOTDIRS_SHALLOW

    out = print_tree(root_dir, exclude_patterns, include_dotfiles, dotdirs_shallow , indent = 2)
    if output_file:
        with open(output_file, 'w', encoding='utf-8') as f:
            for line in out:
                f.write(f"{line}\n")
if __name__ == '__main__':
    main()
