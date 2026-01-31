PYTHON_VENV = ./runtimes/python-3.14.2/.venv
PYTHON = $(PYTHON_VENV)/bin/python3
PYTHON_PIP = $(PYTHON_VENV)/bin/pip
PYTHON_ACTIVATE = $(PYTHON_VENV)/bin/activate

.PHONY: tree activate-venv run

tree:
	@echo "🌳 Proje Ağacı Çıkarılıyor..."
	$(PYTHON) scripts/generate_tree.py

activate-venv:
	@echo "Lütfen şu komutu çalıştırın: source $(PYTHON_ACTIVATE)"

# Run the main project (Go server)
run:
	@echo "Running project (Go server)..."
	@go run ./cmd/agent-server
