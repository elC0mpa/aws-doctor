import re

def replace_in_file(filepath, string_val, const_name):
    with open(filepath, 'r') as f:
        content = f.read()
        
    new_content = re.sub(f'"{string_val}"', const_name, content)
    
    with open(filepath, 'w') as f:
        f.write(new_content)

replace_in_file('service/report/service_waste.go', 'Size', 'colSize')
replace_in_file('service/report/service_waste.go', 'Status', 'colStatus')

# For utils/waste_table/waste_table.go, I'll add the constant definition and then replace.
filepath = 'utils/waste_table/waste_table.go'
with open(filepath, 'r') as f:
    content = f.read()

# Make sure it has colStatus defined
if 'colStatus' not in content:
    lines = content.split('\n')
    for i, line in enumerate(lines):
        if line.startswith('const colEstCostMo'):
            lines.insert(i+1, '\nconst colStatus = "Status"\n')
            break
    content = '\n'.join(lines)

new_content = re.sub(f'"Status"', 'colStatus', content)

with open(filepath, 'w') as f:
    f.write(new_content)

