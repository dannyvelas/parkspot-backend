import re

permit_lps = set()
car_lps = set()
with open('./.prodmigrations/000005_seed_permit.up.sql', 'r') as pfile_in:
    amt_matches = 0
    for line in pfile_in:
        match = re.findall(r"= '([^']+)'", line)
        if match:
            first_match = match[0]
            permit_lps.add(first_match)
            amt_matches += 1
    print(amt_matches)
with open('./.prodmigrations/000003_seed_car.up.sql', 'r') as cfile_in:
    amt_matches = 0
    for line in cfile_in:
        match = re.findall(r", '([^']+)'", line)
        if match:
            first_match = match[0]
            car_lps.add(first_match)
            amt_matches += 1
    print(amt_matches)

amt_not_in_car_lps = 0
amt_here = 0
for lp in permit_lps:
    if lp not in car_lps:
        amt_not_in_car_lps += 1
        print(lp)
    else:
        amt_here += 1
print('amt not of lps not in car.up.sql:', amt_not_in_car_lps)
print(f'amt_here: {amt_here}')
