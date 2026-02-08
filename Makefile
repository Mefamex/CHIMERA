PYTHON_VENV = ./runtimes/python-3.14.2/.venv
PYTHON = $(PYTHON_VENV)/bin/python3
PYTHON_PIP = $(PYTHON_VENV)/bin/pip
PYTHON_ACTIVATE = $(PYTHON_VENV)/bin/activate

.PHONY: help run logs build python-req activate-venv tree tree2 treeA file_analysis stacks_info guiLogs

help:
	@echo "Available helper commands:\n"
	@echo "  make help          - show this help message\n"
	@echo "  make run           - run the main project (Go server)\n"
	@echo "  make logs          - run the log viewer (Go server)\n"
	@echo "  make build         - run build script\n"
	@echo "  make python-req    - install python requirements into venv\n"
	@echo "  make activate-venv - print command to activate python venv\n"
	@echo "  make tree          - generate project tree structure\n"
	@echo "  make tree2         - generate project tree structure (with Go and ASCII)\n"
	@echo "  make treeA         - generate project tree structure (Go & ASCII)\n"
	@echo "  make file_analysis - analyze project files\n"
	@echo "  make stacks_info   - get info about supported stacks\n"
	@echo "  make guiLogs       - run the log inspector (Go GUI)\n"

# Run the main project (Go server)
run:
	@$(MAKE) treeA
	@echo "Running project : ./cmd/agentServer/main.go"
	@go run ./cmd/agentServer

logs:
	@$(MAKE) treeA
	@echo "Log Viewer başlatılıyor..."
	@go run cmd/logViewer/main.go

build:
	@$(MAKE) treeA
	@echo "Running project build script..."
	@bash scripts/build.sh

python-req:
	@echo "🌐 Python Gereksinimler Yükleniyor..."
	$(PYTHON) -m pip install -r runtimes/python-3.14.2/requirements.txt

activate-venv:
	@echo "Lütfen şu komutu çalıştırın: source $(PYTHON_ACTIVATE)"

tree:
	@echo "🌳 Proje Ağacı Çıkarılıyor (py & icon)..."
	$(PYTHON) scripts/generate_tree.py
	@sleep 0.1

tree2:
	@echo "🌳 Proje Ağacı Çıkarılıyor (Go & ASCII)..."
	go run ./scripts/custom_tree.go
	@sleep 0.1

treeA:
	@echo "🌳 Proje Ağacı Çıkarılıyor (FULL)..."
	$(MAKE) file_analysis
	$(MAKE) tree2
	$(MAKE) tree

file_analysis:
	@echo "🌳 Proje Dosyaları Analiz Ediliyor..."
	$(PYTHON) ./scripts/file_analyzer.py
	sleep 0.1

guiLogs:
	@echo "🖥️  Log Inspector Başlatılıyor..."
	@cd cmd/log-gui && go run .