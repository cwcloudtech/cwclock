# AI instruction 32

## Detailed report

In the edit form, an admin or owner should also be able to change the client/project affectation.

## Bug in the report/invoice generation

It seems that an owner cannot see himself in the report or invoice generation even if the corresponding user is selected.

It's probably because a owner is not necessarily a member, there's directly an `user_id` column in the organization.

Moreover, user can be excluded from an organization BUT needs still to be counted in the reports or invoices.
