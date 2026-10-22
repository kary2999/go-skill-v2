package main

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Skill 列表 —— description 对应 SKILL.md / .mdc 的 frontmatter，便于 UI 展示
type Skill struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Triggers    []string `json:"triggers"`  // 触发场景
	Reference   string   `json:"reference"` // 对应 standards/xxx.md（hardcoded/ref-auto）或磁盘绝对路径（gsd-framework）
	Scope       string   `json:"scope"`     // global / conditional
	Source      string   `json:"source"`    // hardcoded / ref-auto / gsd-framework (v1.7.40)
	Group       string   `json:"group"`     // UI 分组标签（v1.7.40）
}

type Demo struct {
	ID          string   `json:"id"`
	File        string   `json:"file"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
}

type Catalog struct {
	Version   string   `json:"version"`
	Skills    []Skill  `json:"skills"`
	Demos     []Demo   `json:"demos"`
	Glossary  []string `json:"glossary"` // 一级分组标题列表
}

var skills = []Skill{
	{
		ID: "iron-laws", Title: "铁律", Description: "团队无条件执行的 14 条硬约束（密钥/金额/时间/错误/并发/日志/SQL 等）",
		Triggers: []string{"所有 Go 文件", "所有 SQL", "提交 commit 时"},
		Reference: "go-style.md", Scope: "always",
	},
	{
		ID: "go-style", Title: "Go 编码风格", Description: "项目结构（Gin 分层 handler → biz → data）、命名、并发、错误、注释、golangci-lint",
		Triggers: []string{"**/*.go"},
		Reference: "go-style.md", Scope: "conditional",
	},
	{
		ID: "naming-logging", Title: "命名 & 日志", Description: "仓库/包/文件命名 + Kafka topic / Redis key 强制规范 + slog 结构化字段 + W3C trace 上下文",
		Triggers: []string{"所有代码", "定义 topic / key", "打日志"},
		Reference: "naming-logging.md", Scope: "always",
	},
	{
		ID: "error-codes", Title: "错误码体系", Description: "业务码 {SERVICE}_{MODULE}_{NNN} + AppError + errcode 包；先登记后使用，CI 校验",
		Triggers: []string{"return error 时"},
		Reference: "error-codes.md", Scope: "conditional",
	},
	{
		ID: "database", Title: "MySQL 设计", Description: "库 / 表命名 / 必备字段（create_time 毫秒 + created_at）/ 自增 id + UUIDv7 oid / 无软删除 / 索引 / 分区 / Migration / 安全",
		Triggers: []string{"migrations/**", "**/*.sql"},
		Reference: "database.md", Scope: "conditional",
	},
	{
		ID: "testing", Title: "单元测试", Description: "覆盖率门禁 / 表驱动 / mockgen / 命名 Test{被测}_{场景}_{期望}",
		Triggers: []string{"**/*_test.go"},
		Reference: "testing.md", Scope: "conditional",
	},
	{
		ID: "commit", Title: "Commit Message", Description: "Conventional Commits + scope 枚举 + commitlint CI 校验",
		Triggers: []string{"写 commit message 时"},
		Reference: "commit.md", Scope: "manual",
	},
	{
		ID: "ci-pipeline", Title: "CI Pipeline（参考）", Description: "六阶段 validate→build→test→scan→package→deploy + MR 门禁 + 性能目标",
		Triggers: []string{"需要了解 CI 原理时"},
		Reference: "ci-pipeline.md", Scope: "manual",
	},
	{
		ID: "common-lib", Title: "优先用 common-lib", Description: "禁止绕过 gin-go-common-lib 直连底层库（kafka-go / go-redis / gorm.Open / c.JSON …），一张替换表",
		Triggers: []string{"**/*.go"},
		Reference: "glossary.md", Scope: "conditional",
	},
	{
		ID: "security", Title: "Cursor / AI 安全红线", Description: ".cursorignore 隔离、输入脱敏、Privacy Mode、禁用自动终端、资金双重 Review",
		Triggers: []string{"所有场景"},
		Reference: "cursor-usage.md", Scope: "always",
	},
	{
		ID: "feature-flags", Title: "特性开关 & 分支管理",
		Description: "FF SDK 集成 + IDP 注册 + GitOps 分支策略 + 5-to-4 预览环境 + Conventional Commits + 治理审计",
		Triggers: []string{"特性开关 / feature flag", "灰度 / 发布 / 回滚", "PR / MR / 分支策略"},
		Reference: "feature-flags.md", Scope: "conditional",
	},
	{
		ID: "id-generation", Title: "订单号 / 分布式 ID",
		Description: "每表自增 id（仅库内）+ oid（UUIDv7 业务 ID / 订单号）；禁止雪花 / 顺序号；idgen 封装；引用用 {entity}_oid",
		Triggers:  []string{"生成订单号 / ID", "定义主键", "**/*.go"},
		Reference: "id-generation.md", Scope: "conditional",
	},
	{
		ID: "api-doc", Title: "接口规范（REST + JSON）",
		Description: "REST 方法与 URL、统一响应体 {code,msg,data,trace_id,timestamp}、HTTP 状态码映射、分页 / 游标、幂等、接口文档模板",
		Triggers:  []string{"写接口 / handler", "接口文档", "前后端联调"},
		Reference: "api-doc.md", Scope: "conditional",
	},
	{
		ID: "redis", Title: "Redis 使用规范",
		Description: "Key 命名（naming.RedisKey）、Cache-Aside、TTL 抖动、分布式锁、幂等第一道防线、大 Key / 热 Key",
		Triggers:  []string{"Redis / 缓存 / 锁", "**/*cache*.go"},
		Reference: "redis.md", Scope: "conditional",
	},
	{
		ID: "kafka", Title: "Kafka 使用规范",
		Description: "Topic 命名、统一消息信封、Outbox、幂等消费（去重表）、handler 内重试 + 死信",
		Triggers:  []string{"Kafka / 事件 / 消费", "Outbox"},
		Reference: "kafka.md", Scope: "conditional",
	},
	{
		ID: "frontend", Title: "前端开发规范",
		Description: "前后端契约分工、请求封装、金额 / 时间处理、错误码文案、安全、发布（不限定框架）",
		Triggers:  []string{"前端 / 页面 / 联调", "请求封装"},
		Reference: "frontend.md", Scope: "conditional",
	},
}

// ============================================================
// 动态扫描 references/ 找未覆盖的 .md，自动生成 ☁️ 卡片
// ============================================================
//
// 设计目的（v1.7.36）：解决「云同步推新规范但 UI 卡片不增加」的问题。
//
// 工作流：
//  1. 启动时（或每次 /api/catalog 请求）扫 ~/.claude/skills/go-team-standards/references/
//  2. 找出**没在硬编码 skills slice 里**的 .md 文件
//  3. 跳过明显的 example/template 文件
//  4. 读 frontmatter title 自动生成卡片，加 ☁️ 前缀让用户识别是云端来源
//
// 副作用：每次 catalog 请求都做磁盘 I/O（references 文件少，开销忽略）

// 已被 hardcoded skills 覆盖的 reference 文件
func hardcodedReferenceFiles() map[string]bool {
	m := map[string]bool{}
	for _, s := range skills {
		m[s.Reference] = true
	}
	return m
}

// 扫 references/ 找未覆盖的 .md，生成 fallback Skill 条目
func scanExtraReferences() []Skill {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	refDir := filepath.Join(home, ".claude", "skills", "go-team-standards", "references")
	entries, err := os.ReadDir(refDir)
	if err != nil {
		return nil // skill 没装 → 没 extras，返回空
	}
	covered := hardcodedReferenceFiles()
	var extras []Skill

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		if covered[name] {
			continue // 已被 hardcoded 覆盖
		}
		// 跳过用户自定义文件（custom-* 由用户自行管理，不作为云端规范卡片展示）
		lowerName := strings.ToLower(name)
		if strings.HasPrefix(lowerName, "custom-") {
			continue
		}

		fp := filepath.Join(refDir, name)
		title := extractFrontmatterTitle(fp)
		if title == "" {
			// fallback：文件名去 .md 后缀
			title = strings.TrimSuffix(name, ".md")
		}
		id := strings.TrimSuffix(name, ".md")

		extras = append(extras, Skill{
			ID:          id,
			Title:       "☁️ " + title,
			Description: "云同步规范（自动发现 from references/）—— 触发由 AI 按文件 frontmatter / 关键字判断。点卡片看完整原文。",
			Triggers:    []string{"按文件 frontmatter / 关键字匹配"},
			Reference:   name,
			Scope:       "conditional",
		})
	}
	return extras
}

// ============================================================
// 扫 ~/.claude/skills/gsd-* 自动加载 GSD 框架 skill 为卡片（v1.7.40）
// ============================================================

func scanGSDFrameworkSkills() []Skill {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	skillsDir := filepath.Join(home, ".claude", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil
	}

	var out []Skill
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "gsd-") {
			continue
		}
		skillFile := filepath.Join(skillsDir, name, "SKILL.md")
		if _, err := os.Stat(skillFile); err != nil {
			continue
		}
		// 解析 frontmatter `name` 和 `description`（fallback）
		fmName, fmDesc := extractFrontmatterNameDesc(skillFile)
		if fmName == "" {
			fmName = name
		}

		// 中文翻译优先（v1.7.41 加）
		var title, desc string
		if zh, ok := gsdTranslations[name]; ok {
			title = "💪 " + zh.Title
			desc = zh.Desc
		} else {
			// 没翻译就 fallback：gsd-plan-phase → "💪 Plan Phase"
			shortName := strings.TrimPrefix(name, "gsd-")
			shortName = strings.ReplaceAll(shortName, "-", " ")
			words := strings.Fields(shortName)
			for i, w := range words {
				if len(w) > 0 {
					words[i] = strings.ToUpper(w[:1]) + w[1:]
				}
			}
			title = "💪 " + strings.Join(words, " ")
			desc = fmDesc
		}

		out = append(out, Skill{
			ID:          name,
			Title:       title,
			Description: desc,
			Triggers:    []string{"GSD 框架 · 上游 gsd-build/get-shit-done"},
			Reference:   skillFile, // 磁盘绝对路径
			Scope:       "conditional",
			Source:      "gsd-framework",
			Group:       "GSD 框架",
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// scanKnownExtraSkills 扫描 Team Standards 随主安装分发的非 gsd-* skill（如 DevDefender）。
// 已安装则生成卡片，未安装则跳过。
func scanKnownExtraSkills() []Skill {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	skillsDir := filepath.Join(home, ".claude", "skills")
	known := []struct {
		dir   string
		emoji string
		group string
	}{
		{"DevDefender", "🛡️", "需求防御"},
	}

	var out []Skill
	for _, k := range known {
		skillFile := filepath.Join(skillsDir, k.dir, "SKILL.md")
		if _, err := os.Stat(skillFile); err != nil {
			continue // 未安装则不展示
		}
		fmName, fmDesc := extractFrontmatterNameDesc(skillFile)
		if fmName == "" {
			fmName = k.dir
		}
		out = append(out, Skill{
			ID:          strings.ToLower(k.dir),
			Title:       k.emoji + " " + fmName,
			Description: fmDesc,
			Triggers:    []string{"输入 /devdefender 或「启动神盾局」激活"},
			Reference:   skillFile,
			Scope:       "conditional",
			Source:      "team-standards",
			Group:       k.group,
		})
	}
	return out
}

// 解析 frontmatter 拿 name 和 description（支持单行 / 多行 / 引号）
func extractFrontmatterNameDesc(fp string) (name, desc string) {
	f, err := os.Open(fp)
	if err != nil {
		return "", ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inFM := false
	inDescMultiline := false
	var descLines []string

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "---" {
			if !inFM {
				inFM = true
				continue
			}
			break
		}
		if !inFM {
			continue
		}

		// 多行 description 收集
		if inDescMultiline {
			// 缩进的行 → 内容；顶格的 key 行 → 结束
			if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
				descLines = append(descLines, strings.TrimSpace(line))
				continue
			}
			// 否则 fall through 当新 key 处理
			inDescMultiline = false
		}

		// 解析 key: value
		if i := strings.Index(line, ":"); i > 0 {
			k := strings.TrimSpace(line[:i])
			v := strings.TrimSpace(line[i+1:])
			v = strings.Trim(v, `"'`)

			if k == "name" {
				name = v
			} else if k == "description" {
				if v == "" || v == "|" || v == ">" || v == "|-" || v == ">-" {
					inDescMultiline = true
				} else {
					desc = v
				}
			}
		}
	}

	if desc == "" && len(descLines) > 0 {
		desc = strings.Join(descLines, " ")
	}
	// 截断过长 description（卡片放不下）
	if len(desc) > 200 {
		desc = desc[:200] + "…"
	}
	return name, desc
}

// 从 .md frontmatter 提取 title 字段
func extractFrontmatterTitle(fp string) string {
	f, err := os.Open(fp)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inFM := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			if !inFM {
				inFM = true
				continue
			}
			return "" // 走出 frontmatter 未找到 title
		}
		if !inFM {
			continue
		}
		if strings.HasPrefix(line, "title:") {
			t := strings.TrimSpace(strings.TrimPrefix(line, "title:"))
			t = strings.Trim(t, `"'`)
			return t
		}
	}
	return ""
}

var demos = []Demo{
	{ID: "gin-service", File: "gin-service-min.go", Title: "Gin 服务骨架", Description: "最小 main.go：logging + middleware + 统一响应 + 幂等 + MySQL/Redis 初始化 + 优雅退出", Keywords: []string{"gin", "新服务", "main", "骨架", "启动", "handler"}},
	{ID: "apperror", File: "apperror.go", Title: "AppError 错误码", Description: "internal/errcode 定义业务错误码 + biz 返回 + handler FailErr", Keywords: []string{"错误码", "errcode", "apperror", "return error"}},
	{ID: "mysql-migration", File: "mysql-migration.sql", Title: "MySQL Migration 模板", Description: "goose up/down + 自增 id + UUIDv7 oid + create_time/update_time + 索引 + COMMENT", Keywords: []string{"建表", "migration", "goose", "schema", "mysql"}},
	{ID: "mysql-gorm-repo", File: "mysql-gorm-repo.go", Title: "GORM Repo 模式", Description: "ctx 超时 + id/oid + 乐观锁 + 游标分页 + 物理删除 + 1062 处理", Keywords: []string{"gorm", "repo", "crud", "查询", "游标分页", "乐观锁"}},
	{ID: "kafka-producer", File: "kafka-producer.go", Title: "Kafka 生产者", Description: "naming.TopicInProject + 统一信封 + PublishWithHeaders", Keywords: []string{"kafka", "生产者", "publish", "发消息", "topic"}},
	{ID: "kafka-consumer", File: "kafka-consumer.go", Title: "Kafka 消费者", Description: "去重表同事务 + handler 内重试 + 死信（不依赖返回 error）", Keywords: []string{"kafka", "consumer", "消费", "订阅", "死信"}},
	{ID: "outbox-relay", File: "outbox-relay.go", Title: "Outbox relay", Description: "SKIP LOCKED 取数 + 保序发送 + 失败重试", Keywords: []string{"outbox", "本地消息表", "relay", "最终一致"}},
	{ID: "redis-cache", File: "redis-cache.go", Title: "Redis 缓存 & 分布式锁", Description: "Cache-Aside + singleflight + TTL 抖动 + SET NX + Lua 释放", Keywords: []string{"redis", "缓存", "锁", "lock", "击穿"}},
	{ID: "slog-trace", File: "slog-trace.go", Title: "logging 日志", Description: "公共库 logging + trace_id + 字段约定", Keywords: []string{"slog", "日志", "trace_id", "logging"}},
	{ID: "table-test", File: "table-driven-test.go", Title: "表驱动单测", Description: "testify/assert + mockgen + 命名约定", Keywords: []string{"单测", "表驱动", "mockgen", "mock", "测试"}},
}

var glossarySections = []string{
	"Go 语言生态", "构建 / 依赖 / 代码生成", "团队自建组件（gin-go-common-lib）",
	"可观测性", "云原生 / Kubernetes / 服务网格", "CI / CD 与质量门禁",
	"安全 / 合规", "消息队列特有术语", "常见命令 / 工具",
}

func handleCatalog(w http.ResponseWriter, r *http.Request) {
	versionBytes, err := readEmbedFile("VERSION")
	version := "unknown"
	if err == nil {
		version = strings.TrimSpace(string(versionBytes))
	}

	// v1.7.43 调整：规范模块 = 开发规范，纯 16 条
	//   1. hardcoded 团队规范 12 条
	//   2. ☁️ ref-auto（references/ 里未覆盖的 .md，~4 张）
	// GSD 框架 66 张卡**不在此处**——它是独立 skill 体系，走独立 sidebar tab（见 /api/gsd-framework/list）
	allSkills := append([]Skill{}, skills...)
	for i := range allSkills {
		if allSkills[i].Source == "" {
			allSkills[i].Source = "hardcoded"
			allSkills[i].Group = "团队规范"
		}
	}
	extras := scanExtraReferences()
	for i := range extras {
		extras[i].Source = "ref-auto"
		extras[i].Group = "云同步规范"
	}
	allSkills = append(allSkills, extras...)

	// 扫 Team Standards 分发的已安装额外 skill（DevDefender 等）
	extraKnown := scanKnownExtraSkills()
	allSkills = append(allSkills, extraKnown...)

	writeJSON(w, http.StatusOK, Catalog{
		Version: version, Skills: allSkills, Demos: demos, Glossary: glossarySections,
	})
}

func handleChangelog(w http.ResponseWriter, r *http.Request) {
	b, err := readEmbedFile("CHANGELOG.md")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "read changelog", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"content": string(b)})
}

func handleDemo(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	if file == "" || strings.Contains(file, "..") || strings.Contains(file, "/") {
		writeError(w, http.StatusBadRequest, "invalid file", nil)
		return
	}
	b, err := readEmbedFile("demos/" + file)
	if err != nil {
		writeError(w, http.StatusNotFound, "demo not found", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"file": file, "content": string(b),
	})
}

// GET /api/gsd-framework/list
//
// 独立返回 GSD 框架的 66 张卡（不再混进 /api/catalog）。
// v1.7.43 调整：规范模块归规范模块，GSD 归 GSD，互不污染。
func handleGSDFrameworkList(w http.ResponseWriter, r *http.Request) {
	cards := scanGSDFrameworkSkills()
	writeJSON(w, http.StatusOK, map[string]any{
		"cards": cards,
		"count": len(cards),
		"hint":  "GSD 框架独立于团队规范——通过 npx 安装，~/.claude/skills/gsd-* 自动扫描",
	})
}

// GET /api/skill-disk-file?path=<abs path>
// 读 ~/.claude/skills/ 下的任意文件（gsd-framework skill 内容预览用，v1.7.40）
// 安全：路径必须在 ~/.claude/skills/ 内
func handleSkillDiskFile(w http.ResponseWriter, r *http.Request) {
	pathQ := r.URL.Query().Get("path")
	if pathQ == "" {
		writeError(w, http.StatusBadRequest, "path required", nil)
		return
	}
	home, err := os.UserHomeDir()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "home", err)
		return
	}
	abs, err := filepath.Abs(pathQ)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad path", err)
		return
	}
	allowedRoot := filepath.Join(home, ".claude", "skills") + string(filepath.Separator)
	if !strings.HasPrefix(abs, allowedRoot) {
		writeError(w, http.StatusForbidden, "path outside ~/.claude/skills/", nil)
		return
	}
	if strings.Contains(abs, "/..") {
		writeError(w, http.StatusForbidden, "path contains ..", nil)
		return
	}
	b, err := os.ReadFile(abs)
	if err != nil {
		writeError(w, http.StatusNotFound, "read file", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"path":    abs,
		"content": string(b),
	})
}

func handleReference(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Query().Get("file")
	if file == "" || strings.Contains(file, "..") || strings.Contains(file, "/") {
		writeError(w, http.StatusBadRequest, "invalid file", nil)
		return
	}
	b, err := readReferenceFile(file)
	if err != nil {
		writeError(w, http.StatusNotFound, "reference not found", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"file": file, "content": string(b),
	})
}

// readReferenceFile 读规范原文：优先磁盘 references/（安装 / 云同步后的真实内容，
// 保证云同步拉来的新规范、以及 scanExtraReferences 扫出的磁盘文件都能点开），
// 磁盘没有再回退到内嵌种子。file 已由调用方校验不含 .. 和 /。
func readReferenceFile(file string) ([]byte, error) {
	if home, e := os.UserHomeDir(); e == nil {
		for _, base := range []string{
			filepath.Join(home, ".claude", "skills", "go-team-standards", "references"),
			filepath.Join(home, ".cursor", "skills-cursor", "go-team-standards", "references"),
		} {
			if b, e := os.ReadFile(filepath.Join(base, file)); e == nil {
				return b, nil
			}
		}
	}
	return readEmbedFile("standards/" + file)
}
