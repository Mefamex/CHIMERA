```text
📁 CHIMERA
┣━━ .gitignore
┣━━ _folder_analysis_results.txt
┣━━ example.txt
┣━━ go.mod
┣━━ go.sum
┣━━ Makefile
┣━━ README.md
┃
┣━━ 📂 cmd
┃   ┣━━ 📂 agentServer
┃   ┃   ┗━━ main.go
┃   ┣━━ 📂 log-gui
┃   ┃   ┣━━ go.mod
┃   ┃   ┣━━ go.sum
┃   ┃   ┗━━ main.go
┃   ┗━━ 📂 logViewer
┃       ┗━━ main.go
┃
┣━━ 📂 configs
┃   ┣━━ config.yaml
┃   ┗━━ 📂 prompts
┃       ┣━━ browser_prompt.txt
┃       ┣━━ cmd_agent_prompt.txt
┃       ┣━━ critic_prompt.txt
┃       ┣━━ json_formatter.txt
┃       ┣━━ planner_prompt.txt
┃       ┣━━ researcher_prompt.txt
┃       ┣━━ start_prompt.txt
┃       ┗━━ todo_to_json.txt
┃
┣━━ 📂 data
┃   ┣━━ readme_logs.md
┃   ┣━━ system_stacks_info.json
┃   ┣━━ 📂 db
┃   ┗━━ 📂 logs
┃
┣━━ 📂 docs
┃   ┣━━ 1-Notlarım.md
┃   ┣━━ 2-Günlük.md
┃   ┣━━ 3-file-tree.md
┃   ┗━━ 4-folder_tree.txt
┃
┣━━ 📂 internal
┃   ┣━━ 📂 agent
┃   ┃   ┣━━ brain.go
┃   ┃   ┣━━ engine.go
┃   ┃   ┣━━ errors.go
┃   ┃   ┣━━ events.go
┃   ┃   ┣━━ history.go
┃   ┃   ┣━━ prompts.go
┃   ┃   ┗━━ structs.go
┃   ┣━━ 📂 clients
┃   ┃   ┗━━ 📂 ollama
┃   ┃       ┣━━ client.go
┃   ┃       ┣━━ examples.go
┃   ┃       ┗━━ types.go
┃   ┣━━ 📂 common
┃   ┃   ┗━━ stacksInfo.go
┃   ┣━━ 📂 core
┃   ┃   ┣━━ config.go
┃   ┃   ┣━━ init.go
┃   ┃   ┗━━ logger.go
┃   ┣━━ 📂 models
┃   ┃   ┗━━ types.go
┃   ┣━━ 📂 server
┃   ┃   ┣━━ hub.go
┃   ┃   ┣━━ log_sink.go
┃   ┃   ┣━━ server.go
┃   ┃   ┣━━ structs.go
┃   ┃   ┗━━ websocket.go
┃   ┗━━ 📂 skills
┃       ┣━━ readFile.py
┃       ┣━━ runCmd.py
┃       ┣━━ tools.json
┃       ┗━━ useTool.go
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
┃   ┣━━ custom_tree.go
┃   ┣━━ custom_tree.py
┃   ┣━━ download_third_party.sh
┃   ┣━━ file_analyzer.py
┃   ┗━━ generate_tree.py
┗━━ 📂 web
    ┗━━ 📂 static
        ┣━━ app.js
        ┣━━ favicon.png
        ┣━━ index.html
        ┣━━ 📂 third_party
        ┃   ┗━━ readme.md
        ┗━━ 📂 viewLogs
            ┗━━ index.html
```