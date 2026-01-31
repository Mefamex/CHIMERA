PYTHON_VENV = ./runtimes/python-3.14.2/.venv
PYTHON = $(PYTHON_VENV)/bin/python3
PYTHON_PIP = $(PYTHON_VENV)/bin/pip
PYTHON_ACTIVATE = $(PYTHON_VENV)/bin/activate

.PHONY: tree activate-venv run

# Run the main project (Go server)
run:
	@echo "Running project (Go server)..."
	@go run ./cmd/agent-server

activate-venv:
	@echo "Lütfen şu komutu çalıştırın: source $(PYTHON_ACTIVATE)"

tree:
	@echo "🌳 Proje Ağacı Çıkarılıyor..."
	$(PYTHON) scripts/generate_tree.py

stacks_info:
	go run -e "package main; import 'CHIMERA/scripts'; func main() { scripts.GetStacksInfo() }"