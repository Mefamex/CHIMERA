# file path : internal/skills/readFile.py

import sys, json
from pathlib import Path
from dataclasses import dataclass, field, asdict
from typing import List, Optional

@dataclass
class FileResult:
    path: str = ""
    error: str = ""
    total_lines: int = 0
    ended_line: int = 0
    lines: List[str] = field(default_factory=list)


def _process_file(path_str: str, encoder='utf-8') -> dict:
    path = Path(path_str).expanduser()
    result = FileResult(path=str(path))
    
    if not path.exists():
        result.error = "Path does not exist."
        return asdict(result)
    
    if not path.is_file():
        result.error = "Path is not a file."
        return asdict(result)

    try:
        # UTF-8 zorlaması yerine errors='replace' kullanmak daha dayanıklıdır.
        with path.open('r', encoding=encoder, errors='replace') as file:
            result.lines = [line.rstrip('\n') for line in file]
        
        result.total_lines = len(result.lines)
        result.ended_line = len(result.lines)
        
    except PermissionError:  
        result.error = "Permission denied."
    except Exception as e: 
        result.error = f"Unexpected error: {str(e)}"
    
    return asdict(result)

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({"error": "Missing arguments", "total_lines": 0, "path": "", "lines": []}))
        sys.exit(1)

    final_dict = _process_file(sys.argv[1])
    print(json.dumps(final_dict))