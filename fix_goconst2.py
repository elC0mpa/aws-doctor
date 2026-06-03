import re
import os

def replace_in_file(filepath, string_val, const_name):
    with open(filepath, 'r') as f:
        content = f.read()
        
    # Remove the bad const injection from previous script
    # Previous script added: '\nconst {const_name} = "{string_val}"' after package line
    # Let's just find it and move it after the import block.
    # Wait, it replaced "string_val" with const_name. The definition is now:
    # const const_name = const_name
    # Because string_val was replaced!
    
    # Let's revert the changes first using git checkout
    pass

if __name__ == '__main__':
    os.system('git checkout cmd/cost.go cmd/report.go service/costexplorer/mapper.go service/ec2/service.go service/pricing/constants.go service/report/service_waste.go utils/csv_output/cost_mappers.go utils/json_output/json_output.go utils/tui/waste_model.go utils/waste_table/waste_table.go')
    
    def apply_fix(filepath, string_val, const_name):
        with open(filepath, 'r') as f:
            content = f.read()
        
        if const_name in content:
            return
            
        # Find the end of imports
        import_end = -1
        lines = content.split('\n')
        in_import = False
        for i, line in enumerate(lines):
            if line.startswith('import ('):
                in_import = True
            elif in_import and line == ')':
                import_end = i
                in_import = False
                break
            elif line.startswith('import "'):
                import_end = i
                break
                
        if import_end != -1:
            lines.insert(import_end + 1, f'\nconst {const_name} = "{string_val}"')
        else:
            # no imports
            for i, line in enumerate(lines):
                if line.startswith('package '):
                    lines.insert(i+1, f'\nconst {const_name} = "{string_val}"')
                    break
                    
        new_content = '\n'.join(lines)
        new_content = re.sub(f'"{string_val}"', const_name, new_content)
        
        with open(filepath, 'w') as f:
            f.write(new_content)
            
    apply_fix('cmd/cost.go', 'cost', 'cmdCostName')
    apply_fix('cmd/report.go', 'report', 'cmdReportName')
    apply_fix('cmd/report.go', 'cost', 'cmdCostName')
    apply_fix('service/costexplorer/mapper.go', 'Amazon Elastic Compute Cloud - Compute', 'svcEC2Name')
    apply_fix('service/costexplorer/mapper.go', 'Amazon Simple Storage Service', 'svcS3Name')
    apply_fix('service/costexplorer/mapper.go', 'AWS Lambda', 'svcLambdaName')
    apply_fix('service/ec2/service.go', 'self', 'ownerSelf')
    apply_fix('service/pricing/constants.go', 'ml.t2.medium', 'instanceMLT2Medium')
    apply_fix('service/pricing/constants.go', 'ml.m5.xlarge', 'instanceMLM5Xlarge')
    apply_fix('service/report/service_waste.go', 'Est. Cost', 'colEstCost')
    apply_fix('service/report/service_waste.go', 'Idle', 'statusIdle')
    apply_fix('service/report/service_waste.go', 'Unused', 'statusUnused')
    apply_fix('utils/csv_output/cost_mappers.go', 'Total Costs', 'strTotalCosts')
    apply_fix('utils/json_output/json_output.go', 'USD', 'currencyUSD')
    apply_fix('utils/tui/waste_model.go', 'EOF', 'statusEOF')
    apply_fix('utils/waste_table/waste_table.go', 'Est. Cost/Mo', 'colEstCostMo')
    apply_fix('utils/waste_table/waste_table.go', 'EXPIRING SOON', 'statusExpiringSoon')
    apply_fix('utils/waste_table/waste_table.go', 'IAM', 'serviceIAM')
    
    os.system('gofumpt -w .')
