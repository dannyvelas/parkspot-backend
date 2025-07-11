import csv
import uuid
import random
import sys
import string
from datetime import datetime, timedelta
from typing import List, Tuple, Union, TypeVar

random.seed(42)

T = TypeVar('T', str, int)


def nullable_to_sql(value: Union[T, None]) -> str:
    if value is None:
        return 'NULL'
    elif isinstance(value, str):
        return f"'{value}'"
    elif isinstance(value, int):
        return f"{value}"


def nullable_to_csv(value: Union[T, None]) -> str:
    if value is None:
        return ''
    else:
        return f'{value}'


def get_rand_tss() -> Tuple[int, int]:
    year = 2023
    month = 2

    rand_month = month + random.randrange(-1, 3)

    rand_day = random.randrange(1, 29)

    start_date = datetime(year, rand_month, rand_day)
    end_date = start_date + timedelta(days=random.randrange(1, 16))

    return (int(start_date.timestamp()), int(end_date.timestamp()))


def get_rand_line(file_name: str) -> str:
    with open(file_name, 'r') as in_file:
        random_line = next(in_file)
        for i, line in enumerate(in_file, 2):
            if random.randrange(i) == 0:
                random_line = line
        return random_line


########################################
# Car
########################################


class Car:
    id: str
    resident_id: str
    license_plate: str
    color: str
    make: Union[str, None]
    model: Union[str, None]
    amt_parking_days_used: int

    def __init__(self, id: str, resident_id: str, license_plate: str, color: str, make: Union[str, None], model: Union[str, None], amt_parking_days_used: int):
        self.id = id
        self.resident_id = resident_id
        self.license_plate = license_plate
        self.color = color
        self.make = make
        self.model = model
        self.amt_parking_days_used = amt_parking_days_used

    def as_sql(self) -> str:
        return (f"INSERT INTO car(id, resident_id, license_plate, color, make, model, amt_parking_days_used) VALUES"
                f"( '{self.id}'"
                f", '{self.resident_id}'"
                f", '{self.license_plate}'"
                f", '{self.color}'"
                f", {nullable_to_sql(self.make)}"
                f", {nullable_to_sql(self.model)}"
                f", {self.amt_parking_days_used}"
                f");"
                )

    def as_csv(self) -> str:
        return (
            f"{self.id}"
            f"\t{self.resident_id}"
            f"\t{self.license_plate}"
            f"\t{self.color}"
            f"\t{nullable_to_csv(self.make)}"
            f"\t{nullable_to_csv(self.model)}"
            f"\t{self.amt_parking_days_used}"
        )


def get_rand_car(resident_id: str) -> Car:
    def get_rand_color() -> str:
        colors = ['gray', 'black', 'blue', 'silver', 'orange',
                  'pink', 'brown', 'purple', 'red', 'green']
        return random.choice(colors)

    def get_rand_lp() -> str:
        return ''.join([random.choice(string.ascii_uppercase + string.digits) for _ in range(random.randrange(6, 9))])

    line = get_rand_line('./scripts/db/gen/csv_in/car.csv')
    split_line = line.split('\t')

    return Car(
        id=str(uuid.UUID(int=random.getrandbits(128), version=4)),
        resident_id=resident_id,
        license_plate=get_rand_lp(),
        color=get_rand_color(),
        make=split_line[0] if bool(random.getrandbits(1)) else None,
        model=split_line[1] if bool(random.getrandbits(1)) else None,
        amt_parking_days_used=random.randrange(0, 30)
    )


def row_to_car(row: List[str]) -> Car:
    return Car(
        id=row[0],
        resident_id=row[1],
        license_plate=row[2],
        color=row[3],
        make=row[4] if row[4] != '' else None,
        model=row[5] if row[4] != '' else None,
        amt_parking_days_used=int(row[6])
    )

########################################
# Resident
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
        return (f"INSERT INTO resident(id, first_name, last_name, phone, email, password, unlim_days, amt_parking_days_used) VALUES"
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


def csv_in_row_to_resident(row: List[str]) -> Resident:
    def get_rand_resid() -> str:
        id_prefix = ''.join([random.choice(string.digits) for _ in range(7)])
        return random.choice(['T', 'B']) + id_prefix

    return Resident(
        id=get_rand_resid(),
        first_name=row[0],
        last_name=row[1],
        phone=row[3],
        email=row[2],
        password=row[4],
        unlim_days=bool(random.getrandbits(1)),
        amt_parking_days_used=random.randrange(0, 31),
    )


def csv_out_row_to_resident(row: List[str]) -> Resident:
    return Resident(
        id=row[0],
        first_name=row[1],
        last_name=row[2],
        phone=row[3],
        email=row[4],
        password=row[5],
        unlim_days=row[6] == 'True',
        amt_parking_days_used=int(row[7]),
    )

########################################
# PERMIT
########################################


class Permit:
    id: int
    resident_id: str
    car_id: str
    license_plate: str
    color: str
    make: Union[str, None]
    model: Union[str, None]
    start_ts: int
    end_ts: int
    request_ts: Union[int, None]
    affects_days: bool
    exception_reason: Union[str, None]

    def __init__(self, id: int, resident_id: str, car_id: str, license_plate: str, color: str, make: Union[str, None], model: Union[str, None], start_ts: int, end_ts: int, request_ts: Union[int, None], affects_days: bool, exception_reason: Union[str, None]):
        self.id = id
        self.resident_id = resident_id
        self.car_id = car_id
        self.license_plate = license_plate
        self.color = color
        self.make = make
        self.model = model
        self.start_ts = start_ts
        self.end_ts = end_ts
        self.request_ts = request_ts
        self.affects_days = affects_days
        self.exception_reason = exception_reason

    def as_sql(self) -> str:
        escaped_reason = self.exception_reason.replace(
            "'", "''") if self.exception_reason else None
        return (f"INSERT INTO permit(id, resident_id, car_id, license_plate, color, make, model, start_ts, end_ts, request_ts, affects_days, exception_reason) VALUES"
                f"( {self.id}"
                f", '{self.resident_id}'"
                f", '{self.car_id}'"
                f", '{self.license_plate}'"
                f", '{self.color}'"
                f", {nullable_to_sql(self.make)}"
                f", {nullable_to_sql(self.model)}"
                f", {self.start_ts}"
                f", {self.end_ts}"
                f", {nullable_to_sql(self.request_ts)}"
                f", {self.affects_days}"
                f", {nullable_to_sql(escaped_reason)});")

    def as_csv(self) -> str:
        return (
            f"{self.id}"
            f"\t{self.resident_id}"
            f"\t{self.car_id}"
            f"\t{self.license_plate}"
            f"\t{self.color}"
            f"\t{nullable_to_csv(self.make)}"
            f"\t{nullable_to_csv(self.model)}"
            f"\t{self.start_ts}"
            f"\t{self.end_ts}"
            f"\t{nullable_to_csv(self.request_ts)}"
            f"\t{self.affects_days}"
            f"\t{nullable_to_csv(self.exception_reason)}"
        )


def row_to_permit(row: List[str]) -> Permit:
    return Permit(
        id=int(row[0]),
        resident_id=row[1],
        car_id=row[2],
        license_plate=row[3],
        color=row[4],
        make=row[5],
        model=row[6],
        start_ts=int(row[7]),
        end_ts=int(row[8]),
        request_ts=int(row[9]) if row[9] != '' else None,
        affects_days=row[10] == 'True',
        exception_reason=row[11]
    )


def get_rand_permit(i: int, resident_id: str, car: Car) -> Permit:
    def get_rand_sentance() -> str:
        with open('./scripts/db/gen/csv_in/sentances.csv', 'r') as in_file:
            random_line = next(in_file)
            for i, line in enumerate(in_file, 2):
                if random.randrange(i) == 0:
                    random_line = line
            return random_line.replace('\n', '')

    start_ts, end_ts = get_rand_tss()
    request_ts = start_ts - random.randrange(0, 259200)

    return Permit(
        i,
        resident_id,
        car.id,
        car.license_plate,
        car.color,
        car.make,
        car.model,
        start_ts,
        end_ts,
        int(request_ts) if bool(random.getrandbits(1)) else None,
        bool(random.getrandbits(1)),
        get_rand_sentance() if bool(random.getrandbits(1)) else None,
    )

########################################
# Visitor
########################################


class Visitor:
    def __init__(self, id: str, resident_id: str, first_name: str, last_name: str, relationship: str, access_start: int, access_end: int):
        self.id = id
        self.resident_id = resident_id
        self.first_name = first_name
        self.last_name = last_name
        self.relationship = relationship
        self.access_start = access_start
        self.access_end = access_end

    def as_sql(self) -> str:
        return (f"INSERT INTO visitor(id, resident_id, first_name, last_name, relationship, access_start, access_end) VALUES"
                f"( '{self.id}'"
                f", '{self.resident_id}'"
                f", '{self.first_name}'"
                f", '{self.last_name}'"
                f", '{self.relationship}'"
                f", {self.access_start}"
                f", {self.access_end}"
                f");"
                )

    def as_csv(self) -> str:
        return (
            f"{self.id}"
            f"\t{self.resident_id}"
            f"\t{self.first_name}"
            f"\t{self.last_name}"
            f"\t{self.relationship}"
            f"\t{self.access_start}"
            f"\t{self.access_end}"
        )


def get_rand_visitor(resident_id: str) -> Visitor:
    start_ts, end_ts = get_rand_tss()
    line = get_rand_line('./scripts/db/gen/csv_in/resident.csv')
    split_line = line.split('\t')

    return Visitor(
        id=str(uuid.UUID(int=random.getrandbits(128), version=4)),
        resident_id=resident_id,
        first_name=split_line[0],
        last_name=split_line[1],
        relationship='fam/fri' if bool(random.getrandbits(1)
                                       ) else 'contractor',
        access_start=start_ts,
        access_end=end_ts,
    )


def row_to_visitor(row: List[str]) -> Visitor:
    return Visitor(
        id=row[0],
        resident_id=row[1],
        first_name=row[2],
        last_name=row[3],
        relationship=row[4],
        access_start=int(row[5]),
        access_end=int(row[6])
    )

########################################
# MAIN
########################################


if __name__ == '__main__':
    if len(sys.argv) < 2:
        print('usage: python3 scripts/db/gen/test_data.py [ csv | migration ]')
        exit(1)
    file_out = sys.argv[1]
    if file_out not in ['csv', 'migration']:
        print('usage: python3 scripts/db/gen/test_data.py [ csv | migration ]')
        exit(1)

    if file_out == 'csv':
        def csv_in_file_name(
            model: str) -> str: return f'./scripts/db/gen/csv_in/{model}.csv'

        def csv_out_file_name(
            model: str) -> str: return f'./scripts/db/gen/csv_out/{model}.csv'

        amt_permits = 0
        with open(csv_out_file_name('resident'), 'w') as r_file_out:
            with open(csv_out_file_name('car'), 'w') as c_file_out:
                with open(csv_out_file_name('visitor'), 'w') as v_file_out:
                    with open(csv_out_file_name('permit'), 'w') as p_file_out:
                        with open(csv_in_file_name('resident'), 'r') as file_in:
                            reader = csv.reader(file_in, delimiter='\t')

                            for _, row in enumerate(reader):
                                # write resident to csv_out
                                resident = csv_in_row_to_resident(row)
                                r_file_out.write(f'{resident.as_csv()}\n')

                                # write car to csv_out
                                cars: List[Car] = []
                                for _ in range(random.randrange(5)):
                                    car = get_rand_car(resident.id)
                                    cars.append(car)
                                    c_file_out.write(f'{car.as_csv()}\n')

                                # write visitors to csv_out
                                for _ in range(random.randrange(5)):
                                    start_ts, end_ts = get_rand_tss()
                                    visitor = get_rand_visitor(resident.id)
                                    v_file_out.write(f'{visitor.as_csv()}\n')

                                # write permits to csv_out
                                for car in cars:
                                    for _ in range(random.randrange(5)):
                                        permit = get_rand_permit(
                                            amt_permits + 1, resident.id, car)
                                        p_file_out.write(
                                            f'{permit.as_csv()}\n')
                                        amt_permits += 1

    elif file_out == 'migration':
        def migration_in_file_name(
            model: str) -> str: return f'./scripts/db/gen/csv_out/{model}.csv'

        def migration_out_file_name(
            version: int, model: str) -> str: return f'./migrations/00000{version}_seed_{model}.up.sql'

        with open(migration_in_file_name('resident'), 'r') as file_in:
            with open(migration_out_file_name(3, 'resident'), 'w') as file_out:
                reader = csv.reader(file_in, delimiter='\t')
                for _, row in enumerate(reader):
                    resident = csv_out_row_to_resident(row)
                    file_out.write(f'{resident.as_sql()}\n')

        with open(migration_in_file_name('car'), 'r') as file_in:
            with open(migration_out_file_name(4, 'car'), 'w') as file_out:
                reader = csv.reader(file_in, delimiter='\t')
                for _, row in enumerate(reader):
                    car = row_to_car(row)
                    file_out.write(f'{car.as_sql()}\n')

        with open(migration_in_file_name('permit'), 'r') as file_in:
            with open(migration_out_file_name(5, 'permit'), 'w') as file_out:
                reader = csv.reader(file_in, delimiter='\t')

                amt_rows = 0
                for _, row in enumerate(reader):
                    permit = row_to_permit(row)
                    file_out.write(f'{permit.as_sql()}\n')

                    amt_rows += 1

                file_out.write(
                    f'\nALTER SEQUENCE permit_id_seq RESTART WITH {amt_rows+1};\n')

        with open(migration_in_file_name('visitor'), 'r') as file_in:
            with open(migration_out_file_name(6, 'visitor'), 'w') as file_out:
                reader = csv.reader(file_in, delimiter='\t')
                for _, row in enumerate(reader):
                    visitor = row_to_visitor(row)
                    file_out.write(f'{visitor.as_sql()}\n')
