tree:
    @echo "🌳 Proje Ağacı Çıkarılıyor..."
    @./runtimes/python-3.14.2/.venv/bin/python scripts/generate_tree.py

# Print how to activate the venv in the current shell
activate-venv:
    @echo "To activate the project's venv in your current shell run:"
    @echo "  source runtimes/python-3.14.2/.venv/bin/activate"

# Open a Python REPL using the project's venv Python
python-repl:
    @echo "Opening Python REPL from venv..."
    @./runtimes/python-3.14.2/.venv/bin/python

# Start an interactive shell with the venv activated
venv-shell:
    @echo "Starting an interactive shell with venv activated..."
    @bash -ic "source runtimes/python-3.14.2/.venv/bin/activate && exec bash"

# Run the main project (Go server)
run:
    @echo "Running project (Go server)..."
    @go run ./cmd/agent-server

.PHONY: tree activate-venv python-repl venv-shell run
