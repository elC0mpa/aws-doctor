import re
import os

def replace_in_file(filepath, string_val, const_name):
    with open(filepath, 'r') as f:
        content = f.read()
    
    if const_name in content:
        return
        
    # Find package line
    lines = content.split('\n')
    for i, line in enumerate(lines):
        if line.startswith('package '):
            lines.insert(i+1, f'\nconst {const_name} = "{string_val}"')
            break
            
    new_content = '\n'.join(lines)
    # Naive replacement of the exact string with quotes
    # Careful not to replace inside other strings or comments
    # Using regex to replace "string_val" outside of other words
    new_content = re.sub(f'"{string_val}"', const_name, new_content)
    
    with open(filepath, 'w') as f:
        f.write(new_content)

replace_in_file('cmd/cost.go', 'cost', 'cmdCostName')
replace_in_file('cmd/report.go', 'report', 'cmdReportName')
replace_in_file('cmd/report.go', 'cost', 'cmdCostName')
replace_in_file('service/costexplorer/mapper.go', 'Amazon Elastic Compute Cloud - Compute', 'svcEC2Name')
replace_in_file('service/costexplorer/mapper.go', 'Amazon Simple Storage Service', 'svcS3Name')
replace_in_file('service/costexplorer/mapper.go', 'AWS Lambda', 'svcLambdaName')
replace_in_file('service/ec2/service.go', 'self', 'ownerSelf')
replace_in_file('service/pricing/constants.go', 'ml.t2.medium', 'instanceMLT2Medium')
replace_in_file('service/pricing/constants.go', 'ml.m5.xlarge', 'instanceMLM5Xlarge')
replace_in_file('service/report/service_waste.go', 'Est. Cost', 'colEstCost')
replace_in_file('service/report/service_waste.go', 'Idle', 'statusIdle')
replace_in_file('service/report/service_waste.go', 'Unused', 'statusUnused')
replace_in_file('utils/csv_output/cost_mappers.go', 'Total Costs', 'strTotalCosts')
replace_in_file('utils/json_output/json_output.go', 'USD', 'currencyUSD')
replace_in_file('utils/tui/waste_model.go', 'EOF', 'statusEOF')
replace_in_file('utils/waste_table/waste_table.go', 'Est. Cost/Mo', 'colEstCostMo')
replace_in_file('utils/waste_table/waste_table.go', 'EXPIRING SOON', 'statusExpiringSoon')
replace_in_file('utils/waste_table/waste_table.go', 'IAM', 'serviceIAM')

