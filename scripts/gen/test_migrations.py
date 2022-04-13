import csv
import uuid
import random
import sys
import string
from datetime import datetime, timedelta
from typing import List, Tuple, Union

random.seed(42)

def nullablestr_to_sql(value: Union[str,None]) -> str:
    if value:
        return f"'{value}'"
    else:
        return 'NULL'

def nullablestr_to_csv(value: Union[str, None]) -> str:
    if value:
        return value
    else:
        return 'NULL'

########################################
## PERMIT
########################################
class Permit:
    id: int
    resident_id: str
    car_id: str
    start_date: str
    end_date: str
    request_ts: int
    affects_days: bool
    def __init__(self, id: int, resident_id: str, car_id: str, start_date: str, end_date: str, request_ts: int, affects_days: bool):
        self.id = id
        self.resident_id = resident_id
        self.car_id = car_id
        self.start_date = start_date
        self.end_date = end_date
        self.request_ts = request_ts
        self.affects_days = affects_days

    def as_sql(self) -> str:
        return (f"INSERT INTO permits(id, resident_id, car_id, start_date, end_date, request_ts, affects_days) VALUES"
            f"( {self.id}"
            f", '{self.resident_id}'"
            f", '{self.car_id}'"
            f", '{self.start_date}'"
            f", '{self.end_date}'"
            f", {self.request_ts}"
            f", {self.affects_days}"
            f");")

    def as_csv(self) -> str:
        return (
            f"{self.id}"
            f"\t{self.resident_id}"
            f"\t{self.car_id}"
            f"\t{self.start_date}"
            f"\t{self.end_date}"
            f"\t{self.request_ts}"
            f"\t{self.affects_days}"
            )

def row_to_permit(row: List[str]) -> Permit:
    return Permit(
        id           = int(row[0]),
        resident_id  = row[1],
        car_id       = row[2],
        start_date   = row[3],
        end_date     = row[4],
        request_ts   = int(row[5]),
        affects_days = row[6] == 'True'
    )

def get_rand_permit(i: int, resident_id: str, car_id: str) -> Permit:
    def get_rand_dates() -> Tuple[datetime, datetime]:
        year = datetime.now().year
        month = datetime.now().month

        rand_year = year - random.randrange(0, 1)
        rand_month = month - random.randrange(0, 3)
        if rand_month < 0:
            rand_month = month

        rand_day = random.randrange(1, 29)

        start_date = datetime(rand_year, rand_month, rand_day)
        end_date = start_date + timedelta(days=random.randrange(1, 16))

        return (start_date, end_date)

    start_date, end_date = get_rand_dates()
    request_ts = start_date.timestamp() - random.randrange(0, 259200)

    return Permit(i, resident_id, car_id, start_date.strftime("%Y-%m-%d"),
            end_date.strftime("%Y-%m-%d"), int(request_ts), bool(random.getrandbits(1)))
########################################
## Car
########################################
class Car:
    id: str
    license_plate: str
    color: str
    make: Union[str, None]
    model: Union[str, None]
    def __init__(self, id: str, license_plate: str, color: str, make: Union[str, None], model: Union[str, None]):
        self.id = id
        self.license_plate = license_plate
        self.color = color
        self.make = make
        self.model = model

    def as_sql(self) -> str:
        return (f"INSERT INTO cars(id, license_plate, color, make, model) VALUES"
            f"( '{self.id}'"
            f", '{self.license_plate}'"
            f", '{self.color}'"
            f", {nullablestr_to_sql(self.make)}"
            f", {nullablestr_to_sql(self.model)}"
            f");"
        )

    def as_csv(self) -> str:
        return (
            f"{self.id}"
            f"\t{self.license_plate}"
            f"\t{self.color}"
            f"\t{nullablestr_to_csv(self.make)}"
            f"\t{nullablestr_to_csv(self.model)}"
        )

def get_rand_car() -> Car:
    def get_rand_line() -> str:
        with open('./scripts/sample_car_data.csv', 'r') as in_file:
            random_line = next(in_file)
            for i, line in enumerate(in_file, 2):
                if random.randrange(i) == 0:
                    random_line = line
            return random_line

    def get_rand_color() -> str:
        colors = ['gray', 'black', 'blue', 'silver', 'orange', 'pink', 'brown', 'purple', 'red', 'green']
        return random.choice(colors)

    def get_rand_lp() -> str:
        return ''.join([random.choice(string.ascii_uppercase + string.digits) for _ in range(random.randrange(6, 9))])

    line = get_rand_line()
    split_line = line.split('\t')

    return Car(
        id            = str(uuid.UUID(int=random.getrandbits(128))),
        license_plate = get_rand_lp(),
        color         = get_rand_color(),
        make          = split_line[0] if bool(random.getrandbits(1)) else None,
        model         = split_line[1] if bool(random.getrandbits(1)) else None
        )

def row_to_car(row: List[str]) -> Car:
    return Car(
      id            = row[0],
      license_plate = row[1],
      color         = row[2],
      make          = row[3],
      model         = row[4]
    )

########################################
## Resident
########################################
class Resident:
    def __init__(self, id: str, first_name: str, last_name: str, phone: str, email: str, password: str, unlim_days: bool, amt_parking_days_used: int):
        self.id = id.upper()
        self.first_name = first_name
        self.last_name = last_name
        self.phone = phone
        self.email = email
        self.password = password
        self.unlim_days = unlim_days
        self.amt_parking_days_used = amt_parking_days_used

    def as_sql(self) -> str:
        return (f"INSERT INTO residents(id, first_name, last_name, phone, email, password, unlim_days, amt_parking_days_used) VALUES"
            f"( '{self.id}'"
            f", '{self.first_name}'"
            f", '{self.last_name}'"
            f", '{self.phone}'"
            f", '{self.email}'"
            f", '{self.password}'"
            f", {self.unlim_days}"
            f", {self.amt_parking_days_used}"
            f");"
        )

    def as_csv(self) -> str:
        return (
            f"{self.id}"
            f"\t{self.first_name}"
            f"\t{self.last_name}"
            f"\t{self.phone}"
            f"\t{self.email}"
            f"\t{self.password}"
            f"\t{self.unlim_days}"
            f"\t{self.amt_parking_days_used}"
        )

def row_to_resident(row: List[str]) -> Resident:
    def get_rand_resid() -> str:
        id_prefix = ''.join([ random.choice(string.digits) for _ in range(7) ])
        return random.choice(['T', 'B']) + id_prefix

    return Resident(
      id                    = get_rand_resid(),
      first_name            = row[0],
      last_name             = row[1],
      phone                 = row[3],
      email                 = row[2],
      password              = row[4],
      unlim_days            = bool(random.getrandbits(1)),
      amt_parking_days_used = random.randrange(0, 31),
        )

########################################
## MAIN
########################################

if __name__ == '__main__':
    if len(sys.argv) < 2:
        print('usage: python3 scripts/gen/test_migrations.py [ csv | migration ]')
        exit(1)
    file_out = sys.argv[1]
    if file_out not in ['csv', 'migration']:
        print('usage: python3 scripts/gen/test_migrations.py [ csv | migration ]')
        exit(1)

    if file_out == 'csv':
        def csv_in_file_name(model: str) -> str: return f'./scripts/gen/csv_in/{model}.csv'
        def csv_out_file_name(model: str) -> str: return f'./scripts/gen/csv_out/{model}.csv'

        with open(csv_out_file_name('residents'), 'w') as r_file_out:
            with open(csv_out_file_name('cars'), 'w') as c_file_out:
                with open(csv_out_file_name('permits'), 'w') as p_file_out:
                    with open(csv_in_file_name('residents'), 'r') as file_in:
                        reader = csv.reader(file_in, delimiter='\t')
                        permit_id = 1

                        for _, row in enumerate(reader):
                            resident = row_to_resident(row)
                            r_file_out.write(f'{resident.as_csv()}\n')

                            car = get_rand_car()
                            c_file_out.write(f'{car.as_csv()}\n')

                            for _ in range(random.randrange(5)):
                                permit = get_rand_permit(permit_id, resident.id, car.id)
                                p_file_out.write(f'{permit.as_csv()}\n')
                                permit_id += 1

    elif file_out == 'migration':
        def migration_in_file_name(model: str) -> str: return f'./scripts/gen/csv_out/{model}.csv'
        def migration_out_file_name(version: int, model: str) -> str: return f'./migrations/00000{version}_seed_{model}.up.sql'

        with open(migration_in_file_name('residents'), 'r') as file_in:
            with open(migration_out_file_name(4, 'residents'), 'w') as file_out:
                reader = csv.reader(file_in, delimiter='\t')
                for _, row in enumerate(reader):
                    resident = row_to_resident(row)
                    file_out.write(f'{resident.as_sql()}\n')
                                        
            with open(migration_in_file_name('cars'), 'r') as file_in:
                with open(migration_out_file_name(3, 'cars'), 'w') as file_out:
                    reader = csv.reader(file_in, delimiter='\t')
                    for _, row in enumerate(reader):
                        car = row_to_car(row)
                        file_out.write(f'{car.as_sql()}\n')

            with open(migration_in_file_name('permits'), 'r') as file_in:
                with open(migration_out_file_name(5, 'permits'), 'w') as file_out:
                    reader = csv.reader(file_in, delimiter='\t')
                    for _, row in enumerate(reader):
                        permit = row_to_permit(row)
                        file_out.write(f'{permit.as_sql()}\n')
