#!/usr/bin/env bash
# soft-delete-guard · PostToolUse
# 铁律：本项目不使用软删除，禁止 is_deleted / deleted_at / is_alive 字段
# 来源：database.md §4.4

set -euo pipefail

TOOL_NAME="${TOOL_NAME:-}"
TOOL_INPUT="${TOOL_INPUT:-}"

case "$TOOL_NAME" in
  write_file|edit_file|str_replace_editor|create_file) ;;
  *) exit 0 ;;
esac

FILE_PATH=$(echo "$TOOL_INPUT" | python3 -c "
import sys, json
try:
    d = json.load(sys.stdin)
    print(d.get('path') or d.get('file_path') or d.get('new_path') or '')
except:
    print('')
" 2>/dev/null || echo "")

[[ -z "$FILE_PATH" ]] && exit 0

case "$FILE_PATH" in
  *.sql|*.go|*.proto) ;;
  *) exit 0 ;;
esac

case "$FILE_PATH" in
  */vendor/*) exit 0 ;;
esac

[[ ! -f "$FILE_PATH" ]] && exit 0

# 检测软删除字段（字段定义或结构体中）
HITS=$(grep -niE '\b(is_deleted|deleted_at|is_alive)\b' "$FILE" \
  | grep -v '^\s*//' | head -5 || true)

if [[ -n "$HITS" ]]; then
  echo "🚫 [soft-delete-guard] 铁律违反：本项目不使用软删除，禁止 is_deleted / deleted_at / is_alive"
  echo ""
  echo "文件：$FILE"
  echo "违规行："
  echo "$HITS" | while IFS= read -r line; do echo "  $line"; done
  echo ""
  echo "📖 依据：database.md §4.4 删除策略 —— 允许物理 DELETE（资金类数据先归档），需要保留状态时用 xxx_status"
  exit 1
fi

exit 0
