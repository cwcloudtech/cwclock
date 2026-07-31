# AI instruction 124

## Mobile app design

* For the report screen, the two date picker have to be centered like it's done for the invoice.
* In the settings:
  * _Switch organization_ button: I want the same icon as _Organization_ button
  * Language button, add a translate icon

## Mobile app features

* I want to be able to modify a time record (add an edit icon on the right of the trash icon)
* Invoices: add the various button reupload, send, delete (existing in the frontend) all with actions icons
* Export job: I want only to be able to run existing export job with a play icon
* Be sure that if I modify an organization, it don't erase external links. If it's the case add a `PATCH` endpoint to not erase it and call it in the mobile
