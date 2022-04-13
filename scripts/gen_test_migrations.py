import csv
import uuid
import random
import string
from datetime import datetime, timedelta
from typing import List, Tuple, Union

def nullablestr_to_sql(value: Union[str,None]) -> str:
    if value:
        return f'{value}'
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

    def as_sql(self):
        return (f"INSERT INTO permits(id, resident_id, car_id, start_date, end_date, request_ts, affects_days) VALUES"
            f"( {self.id}"
            f", '{self.resident_id}'"
            f", '{self.car_id}'"
            f", '{self.start_date}'"
            f", '{self.end_date}'"
            f", {self.request_ts}"
            f", {self.affects_days}"
            f");")

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

    def as_sql(self):
        return (f"INSERT INTO cars(id, license_plate, color, make, model) VALUES"
            f"'{self.id}'"
            f", '{self.license_plate}'"
            f", '{self.color}'"
            f", {nullablestr_to_sql(self.make)}"
            f", {nullablestr_to_sql(self.model)}"
            f");"
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
        id            = str(uuid.uuid4()),
        license_plate = get_rand_lp(),
        color         = get_rand_color(),
        make          = split_line[0] if bool(random.getrandbits(1)) else None,
        model         = split_line[1] if bool(random.getrandbits(1)) else None
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

    def as_sql(self):
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

if __name__ == '__main__':
    def out_file_name(number: int, model: str) -> str: return f'./migrations/00000{number}_seed_{model}.up.sql'

    with open(out_file_name(4, 'residents'), 'w') as r_file_out:
        with open(out_file_name(3, 'cars'), 'w') as c_file_out:
            with open(out_file_name(5, 'permits'), 'w') as p_file_out:
                with open('./scripts/sample_res_data.csv', 'r') as file_in:
                    reader = csv.reader(file_in, delimiter='\t')
                    permit_id = 1
                    for _, row in enumerate(reader):
                        resident = row_to_resident(row)
                        r_file_out.write(f'{resident.as_sql()}\n')

                        car = get_rand_car()
                        c_file_out.write(f'{car.as_sql()}\n')

                        for _ in range(random.randrange(5)):
                            permit = get_rand_permit(permit_id, resident.id, car.id)
                            p_file_out.write(f'{permit.as_sql()}\n')
                            permit_id += 1
