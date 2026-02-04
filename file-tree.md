```text
📁 CHIMERA
┣━━ .gitignore
┣━━ _folder_analysis_results.txt
┣━━ file-structure.md
┣━━ file-tree.md
┣━━ go.mod
┣━━ go.sum
┣━━ Makefile
┣━━ README.md
┃
┣━━ 📂 cmd
┃   ┣━━ try.go
┃   ┗━━ 📂 agent-server
┃       ┗━━ main.go
┃
┣━━ 📂 configs
┃   ┣━━ config.yaml
┃   ┗━━ 📂 prompts
┃       ┣━━ browser_prompt.txt
┃       ┣━━ cmd_agent_prompt.txt
┃       ┣━━ critic_prompt.txt
┃       ┣━━ planner_prompt.txt
┃       ┣━━ researcher_prompt.txt
┃       ┣━━ start_prompt.txt
┃       ┗━━ todo_to_json.txt
┃
┣━━ 📂 data
┃   ┣━━ system_stacks_info.json
┃   ┣━━ 📂 db
┃   ┗━━ 📂 logs
┃       ┗━━ readme-logs.md
┃
┣━━ 📂 internal
┃   ┣━━ 📂 agent
┃   ┃   ┣━━ brain.go
┃   ┃   ┗━━ prompts.go
┃   ┣━━ 📂 clients
┃   ┃   ┗━━ 📂 ollama
┃   ┃       ┣━━ chat.go
┃   ┃       ┣━━ client.go
┃   ┃       ┣━━ listModels.go
┃   ┃       ┣━━ status.go
┃   ┃       ┣━━ stream.go
┃   ┃       ┗━━ structs.go
┃   ┣━━ 📂 common
┃   ┣━━ 📂 core
┃   ┃   ┣━━ config.go
┃   ┃   ┣━━ init.go
┃   ┃   ┗━━ logger.go
┃   ┣━━ 📂 models
┃   ┃   ┗━━ types.go
┃   ┗━━ 📂 server
┃       ┣━━ hub.go
┃       ┣━━ server.go
┃       ┣━━ structs.go
┃       ┗━━ websocket.go
┃
┣━━ 📂 runtimes
┃   ┣━━ 📂 node-v25.4.0
┃   ┃   ┣━━ package.json
┃   ┃   ┣━━ pnpm-lock.yaml
┃   ┃   ┣━━ pnpm-workspace.yaml
┃   ┃   ┗━━ 📂 node_modules
┃   ┗━━ 📂 python-3.14.2
┃       ┣━━ requirements.txt
┃       ┗━━ 📂 .venv
┃
┣━━ 📂 scripts
┃   ┣━━ build.sh
┃   ┣━━ download_third_party.sh
┃   ┣━━ file_analyzer.py
┃   ┣━━ generate_tree.py
┃   ┗━━ stacks_info.go
┃
┣━━ 📂 third_party
┃   ┗━━ readme.md
┃
┣━━ 📂 tools
┗━━ 📂 web
    ┗━━ 📂 static
        ┣━━ app.js
        ┣━━ favicon.png
        ┣━━ index.html
        ┗━━ 📂 third_party
            ┗━━ readme.md
```