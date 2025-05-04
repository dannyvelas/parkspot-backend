import csv
import sys
from typing import List, Union, TypeVar, Dict
from datetime import datetime as dt
import os
import uuid

########################################
# HELPERS
########################################


def migration_file_name(version: int, model: str) -> str:
    return f'./.prodmigrations/00000{version}_seed_{model}.up.sql'


T = TypeVar('T', str, int)


def nullable_to_sql(value: Union[T, None]) -> str:
    if value is None:
        return 'NULL'
    else:
        return to_sql(value)


def to_sql(value: T) -> str:
    if isinstance(value, str):
        escaped = value.replace("'", "''")
        return f"'{escaped}'"
    elif isinstance(value, int):
        return f"{value}"


def str_to_ts(s: str) -> int:
    date_obj = dt.strptime(s, "%Y-%m-%d")
    return int(date_obj.timestamp())


def str_to_end_ts(s: str) -> int:
    date_obj = dt.strptime(s + " 23:59:59", "%Y-%m-%d %H:%M:%S")
    return int(date_obj.timestamp())

########################################
# Admin
########################################


class Props:
    def __init__(self, first_name: str, last_name: str, email: str, is_privileged: bool):
        self.first_name = first_name
        self.last_name = last_name
        self.email = email
        self.is_privileged = is_privileged


class Admin:
    def __init__(self, id: str, first_name: str, last_name: str, email: str, password: str, is_privileged: bool):
        self.id = id
        self.first_name = first_name
        self.last_name = last_name
        self.email = email
        self.password = password
        self.is_privileged = is_privileged

    def as_sql(self):
        return (f"INSERT INTO admin"
                f"( id"
                f", first_name"
                f", last_name"
                f", email"
                f", password"
                f", is_privileged"
                f") VALUES"
                f"( {to_sql(self.id)}"
                f", {to_sql(self.first_name)}"
                f", {to_sql(self.last_name)}"
                f", {to_sql(self.email)}"
                f", {to_sql(self.password)}"
                f", {to_sql(self.is_privileged)}"
                f");")


def row_to_admin(row: List[str], id_to_props: Dict[str, Props]) -> Union[Admin, None]:
    if row[0] in id_to_props:
        props = id_to_props[row[0]]
        return Admin(
            id=row[0],
            first_name=props.first_name,
            last_name=props.last_name,
            email=props.email,
            password=row[1],
            is_privileged=props.is_privileged)
    else:
        return None

########################################
# Car
########################################


class Car:
    id: str
    license_plate: str
    color: str
    make: Union[str, None]
    model: Union[str, None]
    amt_parking_days_used: int

    def __init__(self, id: str, license_plate: str, color: str, make: Union[str, None], model: Union[str, None], amt_parking_days_used: int):
        self.id = id
        self.license_plate = license_plate
        self.color = color
        self.make = make
        self.model = model
        self.amt_parking_days_used = amt_parking_days_used

    def as_sql(self):
        return (f"INSERT INTO car(id, license_plate, color, make, model, amt_parking_days_used) VALUES"
                f"( {to_sql(self.id)}"
                f", {to_sql(self.license_plate)}"
                f", {to_sql(self.color)}"
                f", {nullable_to_sql(self.make)}"
                f", {nullable_to_sql(self.model)}"
                f", {to_sql(self.amt_parking_days_used)}"
                f");"
                )


def row_to_car(row: List[str]) -> Car:
    return Car(
        id=str(uuid.uuid4()),
        license_plate=row[0],
        color=row[1],
        make=None,
        model=None,
        amt_parking_days_used=int(row[2])
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

    def as_sql(self):
        return (f"INSERT INTO resident(id, first_name, last_name, phone, email, password, unlim_days, amt_parking_days_used) VALUES"
                f"( {to_sql(self.id)}"
                f", {to_sql(self.first_name)}"
                f", {to_sql(self.last_name)}"
                f", {to_sql(self.phone)}"
                f", {to_sql(self.email)}"
                f", {to_sql(self.password)}"
                f", {to_sql(self.unlim_days)}"
                f", {to_sql(self.amt_parking_days_used)}"
                f");")


def row_to_resident(row: List[str]) -> Resident:
    return Resident(
        id=row[0],
        first_name=row[1],
        last_name=row[2],
        phone=row[3],
        email=row[4],
        password=row[5],
        unlim_days=row[6] == '1',
        amt_parking_days_used=int(row[7]),
    )

########################################
# PERMIT
########################################


class Permit:
    id: int
    resident_id: str
    license_plate: str
    start_ts: int
    end_ts: int
    request_date: Union[str, None]
    affects_days: bool
    exception_reason: Union[str, None]

    def __init__(self, id: int, resident_id: str, license_plate: str, start_ts: int, end_ts: int, request_ts: Union[int, None], affects_days: bool, exception_reason: Union[str, None]):
        self.id = id
        self.resident_id = resident_id.upper()
        self.license_plate = license_plate
        self.start_ts = start_ts
        self.end_ts = end_ts
        self.request_ts = request_ts
        self.affects_days = affects_days
        self.exception_reason = exception_reason

    def as_sql(self):
        return (f"INSERT INTO permit(id, resident_id, car_id, start_ts, end_ts, request_ts, affects_days, exception_reason) VALUES"
                f"( {to_sql(self.id)}"
                f", {to_sql(self.resident_id)}"
                f", (SELECT car.id FROM car WHERE car.license_plate = '{self.license_plate}')"
                f", {to_sql(self.start_ts)}"
                f", {to_sql(self.end_ts)}"
                f", {nullable_to_sql(self.request_ts)}"
                f", {to_sql(self.affects_days)}"
                f", {nullable_to_sql(self.exception_reason)}"
                f");"
                )


def row_to_permit(row: List[str]) -> Permit:
    return Permit(
        id=int(row[0]),
        resident_id=row[1].upper(),
        license_plate=row[2],
        start_ts=str_to_ts(row[3]),
        end_ts=str_to_end_ts(row[4]),
        request_ts=int(row[5]) if row[5] != 'NULL' else None,
        affects_days=True if row[6] == '1' else False,
        exception_reason=row[7] if row[7] != 'NULL' else None
    )

########################################
# VISITOR
########################################


class Visitor:
    resident_id: str
    first_name: str
    last_name: str
    relationship: str
    access_start: int
    access_end: int

    def __init__(self, resident_id: str, first_name: str, last_name: str, relationship: str, access_start: int, access_end: int):
        self.resident_id = resident_id
        self.first_name = first_name
        self.last_name = last_name
        self.relationship = relationship
        self.access_start = access_start
        self.access_end = access_end

    def as_sql(self):
        return (f"INSERT INTO visitor"
                f"( resident_id"
                f", first_name"
                f", last_name"
                f", relationship"
                f", access_start"
                f", access_end"
                f") VALUES"
                f"( {to_sql(self.resident_id)}"
                f", {to_sql(self.first_name)}"
                f", {to_sql(self.last_name)}"
                f", {to_sql(self.relationship)}"
                f", {to_sql(self.access_start)}"
                f", {to_sql(self.access_end)}"
                f");"
                )


def row_to_visitor(row: List[str]) -> Visitor:
    return Visitor(
        # ignore row[0], it has an id field we won't use
        resident_id=row[1].upper(),
        first_name=row[2],
        last_name=row[3],
        relationship=row[4],
        access_start=str_to_ts(row[5]),
        access_end=str_to_end_ts(row[6])
    )


########################################
# MAIN
########################################
allowed_files = ["permit", "car", "resident", "visitor", "admin"]
if len(sys.argv) < 2:
    print(
        f"usage: python3 gen_prod_migrations.py [{' | '.join(allowed_files)}]")
    exit(1)

model = sys.argv[1]
if model not in allowed_files:
    print(
        f"usage: python3 gen_prod_migrations.py [{' | '.join(allowed_files)}]")
    exit(1)


file_name = f'./scripts/db/gen/prod_csv_in/{model}.csv'
if not os.path.isfile(file_name):
    print(f"Error: {file_name} not found")
    exit(1)

with open(file_name, 'r', encoding='latin-1') as file_in:
    reader = csv.reader(file_in, delimiter='\t')
    next(reader)  # skip header
    if model == 'admin':
        id_to_props = {}
        if not id_to_props:
            print(
                'ERROR: id_to_props is empty.'
                ' Refusing to run the admin prod migration as the resulting file would be empty.'
                ' Please update id_to_props')
        else:
            with open(migration_file_name(2, 'admin'), 'w') as file_out:
                for row in reader:
                    admin = row_to_admin(row, id_to_props)
                    if admin is not None:
                        file_out.write(f'{admin.as_sql()}\n')
    elif model == 'car':
        with open(migration_file_name(3, 'car'), 'w') as file_out:
            for row in reader:
                car = row_to_car(row)
                file_out.write(f'{car.as_sql()}\n')
    elif model == 'resident':
        with open(migration_file_name(4, 'resident'), 'w') as file_out:
            for row in reader:
                resident = row_to_resident(row)
                file_out.write(f'{resident.as_sql()}\n')
    elif model == 'permit':
        with open(migration_file_name(5, 'permit'), 'w') as file_out:
            last_id = -1
            for row in reader:
                permit = row_to_permit(row)
                file_out.write(f'{permit.as_sql()}\n')
                last_id = permit.id

            # alter sequence needed since permit ids are auto-incrementing
            file_out.write(
                f'\nALTER SEQUENCE permit_id_seq RESTART WITH {last_id+1};\n')
    elif model == 'visitor':
        with open(migration_file_name(6, 'visitor'), 'w') as file_out:
            for row in reader:
                visitor = row_to_visitor(row)
                file_out.write(f'{visitor.as_sql()}\n')
