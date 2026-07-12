# AI instruction 40

## Time record edit

In the time record edit forms (time record screen and export detailed for admin/owner), there's a bug with the begin and end hours fields: it's not working when one of the field hour/minute/seconds is lower to 10 (i.e `9` not prefixed with `0`).

The application have to handle both format: `12:9:58 and `12:09:58`.

## External connection list

The delete icon should be the same (a trash) used by other delete function and also display the same kind of confirmation popin.

The title should be `h2` like member list.

Add an external connection should automatically save the organization and refresh the list (add a PATCH endpoint for this that will just add the external connection).

## Members of an organization

A member should be revokable by an admin or owner with a delete button and confirm box as usual.
