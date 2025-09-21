#!/usr/bin/env python3

import re

# 读取文件
with open('backend/app/ai_service.go', 'r') as f:
    content = f.read()

# 替换 InfoWithFields 调用
def replace_info_with_fields(match):
    fields_str = match.group(1)
    message = match.group(2)
    
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
        
        new_call = f'utils.Infof("{message}: {", ".join(field_parts)}", {", ".join(value_parts)})'
    else:
        new_call = f'utils.Infof("{message}")'
    
    return new_call

# 替换 ErrorWithFields 调用
def replace_error_with_fields(match):
    fields_str = match.group(1)
    message = match.group(2)
    
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
        
        new_call = f'utils.Errorf("{message}: {", ".join(field_parts)}", {", ".join(value_parts)})'
    else:
        new_call = f'utils.Errorf("{message}")'
    
    return new_call

# 替换 WarnWithFields 调用
def replace_warn_with_fields(match):
    fields_str = match.group(1)
    message = match.group(2)
    
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
        
        new_call = f'utils.Warnf("{message}: {", ".join(field_parts)}", {", ".join(value_parts)})'
    else:
        new_call = f'utils.Warnf("{message}")'
    
    return new_call

# 多行匹配模式
pattern_info = r'utils\.InfoWithFields\(map\[string\]interface{}\{([^}]+)\}, "([^"]+)"\)'
pattern_error = r'utils\.ErrorWithFields\(map\[string\]interface{}\{([^}]+)\}, "([^"]+)"\)'
pattern_warn = r'utils\.WarnWithFields\(map\[string\]interface{}\{([^}]+)\}, "([^"]+)"\)'

# 替换所有匹配
content = re.sub(pattern_info, replace_info_with_fields, content, flags=re.DOTALL)
content = re.sub(pattern_error, replace_error_with_fields, content, flags=re.DOTALL)
content = re.sub(pattern_warn, replace_warn_with_fields, content, flags=re.DOTALL)

# 写回文件
with open('backend/app/ai_service.go', 'w') as f:
    f.write(content)

print("ai_service.go 剩余 WithFields 替换完成")
