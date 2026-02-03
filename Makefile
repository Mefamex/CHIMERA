PYTHON_VENV = ./runtimes/python-3.14.2/.venv
PYTHON = $(PYTHON_VENV)/bin/python3
PYTHON_PIP = $(PYTHON_VENV)/bin/pip
PYTHON_ACTIVATE = $(PYTHON_VENV)/bin/activate

.PHONY: help run build activate-venv tree tree_golang stacks_info

help:
	@echo "Available helper commands:\n"
	@echo "  make help          - show this help message\n"
	@echo "  make run           - run the main project (Go server)\n"
	@echo "  make build         - run build script\n"
	@echo "  make python-req    - install python requirements into venv\n"
	@echo "  make activate-venv - print command to activate python venv\n"
	@echo "  make tree          - generate project tree structure\n"
	@echo "  make stacks_info   - get info about supported stacks\n"

# Run the main project (Go server)
run:
	@echo "Running project (Go server)..."
	@go run ./cmd/agent-server

build:
	@echo "Running project build script..."
	@bash scripts/build.sh

python-req:
	@echo "🌐 Python Gereksinimler Yükleniyor..."
	$(PYTHON) -m pip install -r runtimes/python-3.14.2/requirements.txt

activate-venv:
	@echo "Lütfen şu komutu çalıştırın: source $(PYTHON_ACTIVATE)"

tree:
	@echo "🌳 Proje Ağacı Çıkarılıyor..."
	$(PYTHON) scripts/generate_tree.py

file_analysis:
	@echo "🌳 Proje Ağacı Çıkarılıyor..."
	$(PYTHON) scripts/file_analyzer.py


stacks_info:
	go run -e "package main; import 'CHIMERA/scripts'; func main() { scripts.GetStacksInfo() }"
