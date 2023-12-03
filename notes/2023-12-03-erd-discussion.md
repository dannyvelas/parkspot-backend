## Auditing
* Instead of having fields in entity tables for auditing purposes, it would be better to have a auditing table, where each row is a transaction.
    * For example, right now, the permit table has a field called `affects_days`.
    * It would make more sense if this permit table did not have this field.
    * Instead, we could have a `permit_audit_table` that has 3 points of information:
        * ID of Resident creating permit
        * details of policy with which permit was created
        * permit details
    * This could be something that goes in MongoDB because there's no requirement for this to be normalized or for data to be joined
    * There could be multiple auditing tables (e.g. one for creating users)
## Making Auditing Methods Swappable
* Right now, we have a `storage` layer, which is essentially a data transfer layer.
    * We don't want to have the logic of recording auditing information to be in the business logic layer. We want it to be in the data transfer layer.
    * Our first implementation of the logic or recording auditing information in the data transfer layer might, for example, insert new rows to a postgresql table. 
    * Our next implementation might, for example, create new records in a mongoDB table.
    * Whenever we switch from one implementation to the next, we would want to make sure that clients of the data transfer layer are none the wiser. The contract between the data transfer layer and the users of the data transfer layer should remain the same.

## UI
* When an admin is creating a policy, they'll have 4 dropdowns:
    * The first one is an entity: what entity will have this restriction? (e.g. Resident/Permit/Car)
    * The second one is: what property will be limited? (e.g. days/amt of cars)
    * the third one is: what will be limited to? (e.g. 10, 5)
    * The fourth one is: calendar unit of time (week/month/year)

## Policy Kinds
* Resident
    * Allowance per year: e.g. a Resident can only use max `x` visitor parking permit days a year
        * 0 == unlimited
    * Maximum number of concurrent permits: a resident can only have `x` concurrent permits at a time
    * A resident can only have `x` concurrent owned cars
* Permit
    * A permit can only be `x` days long
    * (backend-validation) permit cannot outlast the lease of a resident

## Running Policies
* We should return all policy failures, instead of just returning the first policy failure
* We need to remember to run all policy kinds, both resident and permit

## Creating Resident Permits
* right now, it is impossible for admins to create decals for resident cars
* when we add this functionality, an admin should be able to select the policies that apply to the decal of that resident
* a decal will similarly have a many-to-many table that associates a decal to a set of policies
* similarly, an admin should be able to create decals with exceptions for residents and associate custom policies for them
