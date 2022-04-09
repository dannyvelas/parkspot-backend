import re

permit_resids = set()
resident_resids = set()
with open('./migrations/000005_seed_permits.up.sql', 'r') as pfile_in:
    for i, line in enumerate(pfile_in):
        match = re.findall(r'(t|T|b|B)([0-9]+)', line)
        if match:
            first_match = match[0]
            res_id = first_match[0] + first_match[1]
            permit_resids.add(res_id)
    with open('./migrations/000004_seed_residents.up.sql', 'r') as rfile_in:
        for line in rfile_in:
            match = re.findall(r'(t|T|b|B)([0-9]+)', line)
            if match:
                first_match = match[0]
                res_id = first_match[0] + first_match[1]
                resident_resids.add(res_id)

for res_id in permit_resids:
    if res_id not in resident_resids:
        print(res_id)
