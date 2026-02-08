# file path : internal/skills/runCmd.py

import sys
import json
import time
import subprocess


def _run_command(command: str) -> dict:
    started = time.time()

    result = {
        "command": command,
        "stdout": "",
        "stderr": "",
        "exit_code": -1,
        "duration_ms": 0,
        "error": "",
    }

    if not command or not command.strip():
        result["error"] = "Command is empty."
        result["exit_code"] = 2
        return result

    try:
        # Bash ile çalıştır: pipe/redirection gibi shell özellikleri beklendiği için shell=True.
        completed = subprocess.run(
            command,
            shell=True,
            executable="/bin/bash",
            capture_output=True,
            text=True,
        )

        result["stdout"] = completed.stdout or ""
        result["stderr"] = completed.stderr or ""
        result["exit_code"] = int(completed.returncode)

    except FileNotFoundError:
        result["error"] = "bash not found at /bin/bash"
        result["exit_code"] = 127
    except Exception as e:
        result["error"] = f"Unexpected error: {str(e)}"
        result["exit_code"] = 1
    finally:
        result["duration_ms"] = int((time.time() - started) * 1000)

    return result


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(
            json.dumps(
                {
                    "command": "",
                    "stdout": "",
                    "stderr": "",
                    "exit_code": 2,
                    "duration_ms": 0,
                    "error": "Missing arguments",
                }
            )
        )
        sys.exit(2)

    out = _run_command(sys.argv[1])
    print(json.dumps(out))

    # Komutun exit kodunu Python process exit kodu olarak da yansıt.
    try:
        sys.exit(int(out.get("exit_code", 1)))
    except Exception:
        sys.exit(1)
