import csv
import sys
from typing import List
import os
import uuid
import random

def get_rand_color() -> str:
    colors = ['gray', 'black', 'blue', 'silver', 'orange', 'pink', 'brown', 'purple', 'red', 'green']
    random_index = random.randrange(len(colors))
    return colors[random_index]

########################################
## PERMIT
########################################
class Permit:
    id: int
    resident_id: str
    license_plate: str
    color_and_model: str
    start_date: str
    end_date: str
    request_date: str
    affects_days: bool
    def __init__(self, id: int, resident_id: str, license_plate: str, color_and_model: str, start_date: str, end_date: str, request_date: str, affects_days: bool):
        self.id = id
        self.resident_id = resident_id
        self.license_plate = license_plate
        self.color_and_model = color_and_model
        self.start_date = start_date
        self.end_date = end_date
        self.request_date = request_date
        self.affects_days = affects_days

    def as_sql(self):
        return f"""INSERT INTO permits(id, resident_id, car_id, start_date, end_date, request_ts, affects_days) VALUES ({self.id}, '{self.resident_id}', (SELECT cars.id FROM cars WHERE cars.license_plate = '{self.license_plate}'), '{self.start_date}', '{self.end_date}', {self.request_date}, {self.affects_days});"""

def row_to_permit(row: List[str]) -> Permit:
    for e in row:
        e = e.replace('"', '')
    return Permit(
            id = int(row[0]),
            resident_id = row[1].upper(),
            license_plate = row[2],
            color_and_model = row[3],
            start_date = row[4],
            end_date = row[5],
            request_date = row[6] if row[6] != 'NULL' else 'NULL',
            affects_days = True if row[7] == '1' else False
            )

########################################
## Car
########################################
class Car:
    id: str
    license_plate: str
    color: str
    make: str
    model: str
    def __init__(self, id: str, license_plate: str, color: str, make: str, model: str):
        self.id = id
        self.license_plate = license_plate
        self.color = color
        self.make = make
        self.model = model
    def as_sql(self):
        return f"""INSERT INTO cars(id, license_plate, color, make, model) VALUES ( '{self.id}', '{self.license_plate}', '{self.color}', '{self.make}', '{self.model}');"""

def row_to_car(row: List[str]) -> Car:
    for e in row:
        e = e.replace('"', '')
    return Car(
        id            = str(uuid.uuid4()),
        license_plate = row[0],
        color         = get_rand_color(),
        make          = 'toyota',
        model         = 'tercel'
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
        return f"""INSERT INTO residents(id, first_name, last_name, phone, email, password, unlim_days, amt_parking_days_used) VALUES ('{self.id}', '{self.first_name}', '{self.last_name}', '{self.phone}', '{self.email}', '{self.password}', {self.unlim_days}, {self.amt_parking_days_used});"""

def row_to_resident(row: List[str]) -> Resident:
    for e in row:
        e = e.replace('"', '')
    return Resident(
      id                    = row[0],
      first_name            = row[1],
      last_name             = row[2],
      phone                 = row[3],
      email                 = row[4],
      password              = row[5],
      unlim_days            = row[6] == '1',
      amt_parking_days_used = int(row[7]),
        )
########################################
## MAIN
########################################
allowed_files = [ "permits", "cars", "residents" ]
if len(sys.argv) < 2:
    print(f"usage: python3 gen_migrations.py [{' | '.join(allowed_files)}]")
    exit(1)
elif sys.argv[1] not in allowed_files:
    print(f"usage: python3 gen_migrations.py [{' | '.join(allowed_files)}]")
    exit(1)

model = sys.argv[1]

file_name = f'./.prodtables/{model}.csv'
if not os.path.isfile(file_name) :
    print(f"Error: {file_name} not found")
    exit(1)

with open(file_name, 'r') as file_in:
    with open(f'./scripts/{model}_out.sql', 'w') as file_out:
        if model == 'permits':
            reader = csv.reader(file_in) # different csv delimiter as others. no header
            for row in reader:
                permit = row_to_permit(row)
                file_out.write(f'{permit.as_sql()}\n')
        if model == 'cars':
            reader = csv.reader(file_in, delimiter='\t')
            next(reader) # skip header
            for row in reader:
                car = row_to_car(row)
                file_out.write(f'{car.as_sql()}\n')
        if model == 'residents':
            reader = csv.reader(file_in, delimiter='\t')
            next(reader) # skip header
            for row in reader:
                resident = row_to_resident(row)
                file_out.write(f'{resident.as_sql()}\n')
