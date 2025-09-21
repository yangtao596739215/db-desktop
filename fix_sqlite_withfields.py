#!/usr/bin/env python3

import re

# 读取文件
with open('backend/sqlite/manager.go', 'r') as f:
    content = f.read()

# 替换 m.logger.WithFields 调用
def replace_logger_with_fields(match):
    fields_str = match.group(1)
    log_level = match.group(2)
    message = match.group(3)
    
    # 解析字段
    fields = {}
    for field_match in re.finditer(r'"([^"]+)":\s*([^,}]+)', fields_str):
        key = field_match.group(1)
        value = field_match.group(2).strip()
        fields[key] = value
    
    # 构建新的日志调用
    if fields:
        field_parts = []
        value_parts = []
        for key, value in fields.items():
            field_parts.append(f"{key}=%s")
            value_parts.append(value)
        
        new_call = f'm.logger.{log_level}f("{message}: {", ".join(field_parts)}", {", ".join(value_parts)})'
    else:
        new_call = f'm.logger.{log_level}f("{message}")'
    
    return new_call

# 多行匹配模式
pattern = r'm\.logger\.WithFields\(logrus\.Fields\{([^}]+)\}\)\.(Info|Error|Warn|Debug)\("([^"]+)"\)'

# 替换所有匹配
content = re.sub(pattern, replace_logger_with_fields, content, flags=re.DOTALL)

# 写回文件
with open('backend/sqlite/manager.go', 'w') as f:
    f.write(content)

print("sqlite/manager.go WithFields 替换完成")
