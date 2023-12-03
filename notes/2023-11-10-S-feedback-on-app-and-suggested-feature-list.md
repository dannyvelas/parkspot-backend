Nov 10, 2023

* Presentation
    * The second slide should present the problem more directly: saying "in communities with more residents' cars than car spaces". The way it is right now implies that it's a problem that there are a lot of residents.
    * The problem that the second evil person is saying is actually solved by using resident decals not by ParkSpot at the moment.

* is easier to sell current customer on new features than getting new customer

| ID| Conceptual | ParkSpot Usage |
|---+------------+----------------|
| 1 | Apartment  | Resident       |
| 2 | Resident   | Does not exist |



* Conceptual Unit
    1 Never deleted
    2 One of these maps to one physical living unit in a community
* Conceptual Resident
    * One is created when someone moves in
    * One is deleted when someone moves out
    * Multiple of these can be related to one conceptutal unit

* Parkspot:
    * Resident (hybrid of the two above)
        * feature no.2 of conceptual unit: it maps to one physical living unit in a community
        * feature no.1 and no.2 of a conceptual resident: one is created when someone moves in and deleted when someone moves out

* Simple
    * create a concept of a tenant
    * the current parkspot definition of a resident (which maps 1-1 a physical unit) will get some amount of parking permits
    * in the current parkspot definition of a resident we will add a column called "no. of rooms"
    * re-name "resident" table to be called "apartment"
        * id (1:123)
        * first name
        * last name
        * phone
        * email
        * password
        * no.of rooms
* Medium alternative
    * create a schema for tenant
    * Create schema for unit
        * id column (building id + apartment id)
        * no. of rooms
    * create schema resident for resident (as defined above)
        * id column (auto increment integer)
        * first name
        * last name
        * phone
        * email
        * password
* Hard alternative
    * Medium and also...
    * Create a concept of a building in schema

* Feature list
    * Nov 12 2023
        ✓ would be good to add feature to unify resident decal creation, this should be something that only admins can do
            * if we do this, we will have to have a distinction for resident cars. For a given resident on ParkSpot, one set of cars will be "My Cars" (e.g. cars that belong to the resident that have a decal); the other set of cars will be "Visitor Cars" (e.g. cars that belong to the visitors of a resident that have been given permits)
        * might be better to use filtering for lists in the frontend instead of tabs 
        ✓ it would be good if admins could change their visitor parking policy on ParkSpot
        ✓ it would be required for ParkSpot to be able to support more than 1 community. it should allow a system administrator to add a new community to onboard.
        ✓ as a part of this, there should be a new user type: System Administrator
            ✓ System administrators should be able to create new communities
            ✓ System administrators should also be able to create an "Admin" accounts
        ✓ "Admin" accounts should be able to CRUD "Security" accounts for their community
        * When a new Resident is created by an "Admin" of community "X", they needs to be associated with the "X".
        * When that Resident creates a visitor, then that visitor should automatically be associated also with the resident
        * Would be good to have a search or auto-complete for a resident
        * Would be good for a resident field to be "remaining balance" instead of `days_used`. As it stands, there would be an inconsistency if you were to count the amount of parking days used by checking permits and the value of the field `days_used` for a resident
            * This would probably be good to refine: should we set up a system where people can increase balance?
            * There should also be logs or traces at least so it can be determined why a person has a given balance at a given time
        * You need to have concept of apartment where apartment is inside building
        * Might be better to use concept of apartment
        * If we had the concept of resident in the app. It would better model real life. In real life, there can be multiple residents for the same apartment. It would be good to record these residents. For example, Joe in 2020 can be renting apartment 123 for 6 months. And Amy in 2023 can be renting apartment 123 for 3 years. If we had the concept of residents in the application, we would be able to see the permits that belonged to Joe and contrast them to the permits that belong to Amy, even though they belong to the same apartment, 123.
            * With this approach, we will keep the concept of resident type but not the concept of apartment type
            * Residents can be linked via relation to a row in the apartment table: one-to-many
                * Residents table will have a foreign key column called apartment ID which will be duplicated every now and then
        * Might be easier for admins to select building from a dropdown and apartment from a dropdown instead of putting resident ID
        * Will be easier for all the residents within one "apartment" at a given time to share a balance. individual residents that live inside an apartment at the same time should not have unique balances.
        * Think about how to make ParkSpot as customizable as possible, instead of hard-coding strict hierarchy rules and policies.
            * So for example, no. of days for permit policy should be customizable. In some communities that are bigger it will be a bigger number, in others it will be smaller
        * It might be a good thing for permits to have a field that indicate who that permit was created by, for auditing purposes
        * make visitors not have repititions
        * make visitors list be searchable by resident name
        * you should be able to search for only a certain field in the database
    * Nov 19 2022
        * it would be good to allow people to have a theme for their application

* Scalability Considerations
    * At minimum, it would make sense to have 1 server for each of these:
        * Landing page (as a CDN)
        * frontend
        * backend
        * database
    * If we have multiple communities, it might make sense to have one database for each community.
        * Having one database for all communities is easier, however it can have performance costs, because the database tables will grow a lot.
        * Also, it means there is a single point of failure, so if the database fails, everything fails.
        * Managing more than one database can be difficult, so a potential solution to make management of multiple databases is easier, is to use database as a service (DAAS).

* Security Considerations
    * `postgres` user should not be used as user that is used by backend to interface with the database. it is a security concern. it would be better to have one user that the backend uses to interface with the database, and have another user that a database administrator uses to interface with the database.
    * it would be better to have a user for each person signing in to the server, instead of using the `root` account
