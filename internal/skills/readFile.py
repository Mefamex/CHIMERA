# file path : internal/skills/readFile.py

import sys, json
from pathlib import Path
from dataclasses import dataclass, field, asdict
from typing import List, Optional

@dataclass
class FileResult:
    path: str = ""
    error: Optional[str] = None  # Hata yoksa None olması daha Pythonic'tir
    total_lines: Optional[int] = None
    ended_line: Optional[int] = None
    lines: List[str] = field(default_factory=list)


def _process_file(path_str: str, encoder='utf-8') -> dict:
    # expanduser to handle paths starting with ~ (home directory)
    path = Path(path_str).expanduser()
    result = FileResult(path=str(path))
    
    if not path.exists():
        result.error = "The provided path does not exist."
        return asdict(result)
    
    if not path.is_file():
        if path.is_dir():  result.error = "The provided path is a directory, not a file."
        else:  result.error = "The provided path is neither a file nor a directory."
        return asdict(result)

    try:
        with path.open('r', encoding=encoder) as file: result.lines = [line.rstrip('\n') for line in file]
        result.total_lines = len(result.lines)
        result.ended_line = len(result.lines)
    except UnicodeDecodeError: result.error = "Dosya kodlama (encoding) hatası. UTF-8 değil."
    except PermissionError:  result.error = "Okuma izni reddedildi."
    except Exception as e: result.error = f"Beklenmedik bir hata: {str(e)}"
    
    return asdict(result)
    """
    if not Path(file_path).exists():
        result["error"] = "The provided file path does not exist. "
        if Path(file_path).is_dir(): result["error"] += "The provided path is a directory, not a file."
        if Path(file_path).parent.exists():
            # write the files in folder but first 5 file...
            result["error"] += "However, the parent directory exists."
            parent_dir = Path(file_path).parent
            files = list(parent_dir.iterdir())
            if files:
                result["error"] += " Here are some files in the parent directory: "
                for f in files[:5]:result["error"] += f.name + ", "
                result["error"] = result["error"].rstrip(", ") + "."
            else: result["error"] += " However, the parent directory is empty."
        else : result["error"] += "The parent directory of the provided file path does not exist."
    elif Path(file_path).is_file():
        try:
            with open(file_path, 'r', encoding='utf-8') as file:
                lines = file.readlines()
                result["path"] = str(file_path)
                result["total_lines"] = len(lines)
                result["srarted_line"] = 0
                result["ended_line"] = len(lines)
                result["lines"] = [line.rstrip('\n') for line in lines]
        except Exception as e: result["error"] = f"An error occurred while reading the file: {str(e)}"
    """
    return result

def ReadFile(path_str,  encoder='utf-8') -> dict:
    return _process_file(path_str, encoder)


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({"error": "Missing arguments", "total_lines": 0}))
        sys.exit(1)

    final_dict = ReadFile(sys.argv[1])
    
    print(json.dumps(final_dict))