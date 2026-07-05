# AI instruction 3

## Profile

The _update picture_ must be in the edit profile modal, not the navbar dropdown.

It must also be renamed _update avatar_.

## Super admin and confirmed user

I want the following global roles directly stored in the user's data payload:
* superuser
* confirmed
* disabled

The first signed up user is automatically set as `superuser`, other must be `disabled` until the `superuser` confirms them.

If the user is `disabled` he can't do anything and the app needs to display that he has to _contact an administrator_.

The `superuser` needs to have a screen to manage and edit all users (including their role but also all others field like name, surname, avatar, password, etc.). Password already set must not be displayed in the edit form but only overridable. The backend needs to check if the token is `superuser` if it's not the right connected user.

The `superuser` can also update/delete any organization and change the ownership (transfering the ownership will be sufficient if he want to update the projects, clients or time records so no need to go further).

## Updating time record

If the user is `owner` or `admin` of the organization, the select field of the user in the time record must be replaced by an autocomplete.
