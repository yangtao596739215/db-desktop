#!/usr/bin/env python3

import re
import os
import glob

# 处理单个文件
def process_file(filepath):
    print(f"处理文件: {filepath}")
    
    with open(filepath, 'r') as f:
        content = f.read()

    # 替换 s.logger.WithFields 调用
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
            
            new_call = f's.logger.{log_level}f("{message}: {", ".join(field_parts)}", {", ".join(value_parts)})'
        else:
            new_call = f's.logger.{log_level}f("{message}")'
        
        return new_call

    # 多行匹配模式
    pattern = r's\.logger\.WithFields\(logrus\.Fields\{([^}]+)\}\)\.(Info|Error|Warn|Debug)\("([^"]+)"\)'

    # 替换所有匹配
    new_content = re.sub(pattern, replace_logger_with_fields, content, flags=re.DOTALL)
    
    # 写回文件
    with open(filepath, 'w') as f:
        f.write(new_content)

# 处理 backend/database/ 目录下的所有 .go 文件
database_files = glob.glob('backend/database/*.go')
for filepath in database_files:
    process_file(filepath)

print("database/ 目录下所有文件的 WithFields 替换完成")
