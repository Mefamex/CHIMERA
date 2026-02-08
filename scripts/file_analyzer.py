# -*- coding: utf-8 -*-
# Created on  : 2025-08-09
# Updated on  : 2026-02-08
# @author     : mefamex
# FOR         : folder depth file

__version__ = '1.0.4'

import os,time, datetime

print_results = False
# Hedef Uzantılar
TARGET_EXTENSIONS = []# {'txt', 'bat', 'md', 'py', 'html', 'css', 'js', 'json', 'webmanifest'}
# starts with or ends with
IGNORED_FOLDERS_NAME = [
    'out', 'build', 'dist', 'bin', 'obj', 'env', 'venv', '.venv', 'node_modules',
    '.git', '.svn', '.hg', '.idea', '.vscode', '__pycache__', 'third_party' , 'node_modules'
]
IGNORED_FOLDERS_STARTS_WITH = ['.', '__']
IGNORED_FOLDERS_ENDS_WITH = ['__pycache__']
IGNORED_FILES_ENDS_WITH = ['.log', '.tmp', '.bak', '.backup', '.swp', '.swo','.png']

PROJECT_NAME = "CHIMERA-AI"  # Path(os.getcwd()).name


text = f'# Created for : {PROJECT_NAME}\n# Created on  : ' + datetime.datetime.now().isoformat(timespec='seconds') + 'Z\n'

def count_characters_in_file(file_path):
    try:
        with open(file_path, 'r', encoding='utf-8') as file:
            content = file.read()
            return len(content.replace(" ", "").replace("\n", "").replace("\t", ""))
    except Exception as e:
        print(f"Hata: {file_path} dosyası okunamadı - {str(e)}")
        return 0



def create_table(results, grand_total_char, grand_total_files):
    """Tablo formatında sonuçları oluşturur"""
    # Verileri hazırla (char sayısına göre sıralanmış)
    table_data, sorted_exts = [], sorted(results.keys(), key=lambda ext: results[ext]['total_chars'], reverse=True)
    for ext in sorted_exts:
        file_count, char_count = len(results[ext]['files']), results[ext]['total_chars']
        if file_count == 0 and char_count == 0: continue
        table_data.append({ 'type': ext.upper(), 'char': f"{char_count:,}" if char_count > 0 else "---", 'file': str(file_count) if file_count > 0 else "---" })
    # Sütun genişliklerini hesapla
    type_width = max(len("TYPE"), max(len(row['type']) for row in table_data)) + 2
    char_width = max(len("CHAR"), max(len(row['char']) for row in table_data)) + 2
    file_width = max(len("FILE"), max(len(row['file']) for row in table_data)) + 2
    # Tablo çizimi
    separator = "|" + "=" * type_width + "|" + "=" * char_width + "|" + "=" * file_width + "|"
    header_line = f"|{'TYPE':^{type_width}}|{'CHAR':^{char_width}}|{'FILE':^{file_width}}|"
    table = f"\nANALYZING RESULTS\n\n{separator}\n{header_line}\n"
    table += "|" + "-" * type_width + "|" + "-" * char_width + "|" + "-" * file_width + "|\n"
    for row in table_data: table += f"|{row['type']:>{type_width-1}} |{row['char']:>{char_width-1}} | {row['file']:<{file_width-1}}|\n"
    table += f"{separator}\n\nTOTAL CHARS : {grand_total_char:,}\nTOTAL FILES : {grand_total_files:,}\n\n"+ f"in {PROJECT_NAME}".rjust(32) + "\n"
    return table

def check_ignored_folders(folder_name):
    ignored_folders_name = IGNORED_FOLDERS_NAME
    ignored_folders_starts_with = IGNORED_FOLDERS_STARTS_WITH
    ignored_folders_ends_with = IGNORED_FOLDERS_ENDS_WITH
    # Gelen değerin yalnızca klasör adını al
    base_name = folder_name.split(os.sep)[-1]
    # Tam eşleşme kontrolü
    if base_name in ignored_folders_name: return True
    # Önek kontrolü (örn. gizli dosyalar: .gitignore gibi)
    for prefix in ignored_folders_starts_with:
        if base_name.startswith(prefix) and base_name != prefix:return True
    # Sonek kontrolü
    for suffix in ignored_folders_ends_with:
        if base_name.endswith(suffix) and base_name != suffix:return True
    return False

def analyze_directory(directory,print_results=True):
    global text
    target_extensions = TARGET_EXTENSIONS
    grand_total_char, grand_total_files = 0, 0
    if not target_extensions:
        results = {}
        accept_all_exts = True
    else:
        results = {ext.lstrip('.'): {'files': [], 'total_chars': 0} for ext in target_extensions}
        accept_all_exts = False
    ignored_ALL = []
    ignored_FILES_ALL = []
    if print_results: print("ANALYZING...")
    for root, dirs, files in os.walk(directory):
        ignored = []
        for d in list(dirs):
            if check_ignored_folders(d):
                ignored.append(d)
                dirs.remove(d)
        for d in ignored:
            rel_path = os.path.relpath(os.path.join(root, d), directory)
            if rel_path == ".": rel_path = "\\"
            else: rel_path = "\\" + rel_path.replace("\\", "/")
            if print_results: print(f"!!! PASS : {rel_path}")
        rel_path = os.path.relpath(root, directory)
        if rel_path == ".": rel_path = "\\"
        else: rel_path = "\\" + rel_path.replace("\\", "/")
        if print_results: print("reading folder:", root)
        # dosya bazlı yoksayılan ekler (sonek ile)
        ignored_files = []
        for file in files:
            # sonek bazlı yoksayma kontrolü
            if any(file.lower().endswith(suffix.lower()) for suffix in IGNORED_FILES_ENDS_WITH):
                ignored_files.append(file)
                if print_results:
                    rel_path = os.path.relpath(os.path.join(root, file), directory)
                    if rel_path == ".": rel_path = "\\"
                    else: rel_path = "\\" + rel_path.replace("\\", "/")
                    if print_results: print(f"!!! PASS FILE : {rel_path}")
                continue

            _, ext = os.path.splitext(file)
            ext_key = ext.lstrip('.')
            if ext_key == '': ext_key = 'noext'
            file_path = os.path.join(root, file)
            char_count = count_characters_in_file(file_path)
            if accept_all_exts:
                if ext_key not in results:  results[ext_key] = {'files': [], 'total_chars': 0}
                results[ext_key]['files'].append(file_path)
                results[ext_key]['total_chars'] += char_count
            else:
                if any(ext_key == key.lstrip('.') for key in target_extensions):
                    results[ext_key]['files'].append(file_path)
                    results[ext_key]['total_chars'] += char_count

        ignored_ALL.extend(ignored)
        ignored_FILES_ALL.extend(ignored_files)

    # rapor amaçlı (kök üzerinde bulunan son tutulan ignored listesini yazdır)
    for ign in ignored_ALL:
        rel_path = os.path.relpath(os.path.join(directory, ign), directory)
        if print_results: print(f"!!! PASS : {rel_path}")
    for ignf in ignored_FILES_ALL:
        rel_path = os.path.relpath(ignf, directory)
        if print_results: print(f"!!! PASS FILE (summary) : {rel_path}")

    # Toplam değerleri hesapla
    for ext in results:
        grand_total_char += results[ext]['total_chars']
        grand_total_files += len(results[ext]['files'])

    # Tablo oluştur
    table_output = create_table(results, grand_total_char, grand_total_files)
    text += table_output

    if print_results: print("\nDONE.\n\n\n")



if __name__ == "__main__":
    current_dir = os.getcwd()
    
    print(f"Analiz edilen dizin: {current_dir}")
    #time.sleep(1)
    
    analyze_directory(current_dir, print_results=print_results)
    
    # Sonuçları yazdır
    for line in text.split("\n"):
        if print_results: print(line)
        #time.sleep(0.1)

    # Dosyaya kaydet
    filename = f"_folder_analysis_results.txt" #_{datetime.datetime.now().strftime('%Y-%m-%d %H.%M.%S')}.txt"
    with open(filename, "w", encoding="utf-8") as f: f.write(text)

    print(f"\n\nAnalysis results have been saved to '{os.path.abspath(filename)}'.")