import re

permit_resids = set()
resident_resids = set()
with open('./.prodmigrations/000005_seed_permit.up.sql', 'r') as pfile_in:
    amt_lines_read = 0
    amt_matches = 0
    for line in pfile_in:
        match = re.findall(r'(T|B)([0-9]+)', line)
        if match:
            first_match = match[0]
            res_id = first_match[0] + first_match[1]
            permit_resids.add(res_id)
            amt_matches += 1
        amt_lines_read += 1
    print(f'amt_lines_read: {amt_lines_read}. amt_matches: {amt_matches}')

with open('./.prodmigrations/000004_seed_resident.up.sql', 'r') as rfile_in:
    amt_lines_read = 0
    amt_matches = 0
    for line in rfile_in:
        match = re.findall(r'(T|B)([0-9]+)', line)
        if match:
            first_match = match[0]
            res_id = first_match[0] + first_match[1]
            resident_resids.add(res_id)
            amt_matches += 1
        amt_lines_read += 1
    print(f'amt_lines_read: {amt_lines_read}. amt_matches: {amt_matches}')

amt_here = 0
for res_id in permit_resids:
    if res_id not in resident_resids:
        print(res_id)
    else:
        amt_here = amt_here + 1

print(f'amt_here: {amt_here}')
