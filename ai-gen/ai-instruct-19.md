# AI instruction 19

## Client and project daily rate

An organization have a daily rate per member, but I want to also define a daily rate on those level:
* on the project
* on the client

If a client daily rate is defined, it's prior to the member daily rate.

If a project daily rate is defined, it's prior to the client daily rate.

It has to be taken into account in the details and summary report and will be used later in the invoicing feature.

## Template refactoring

In the backend, I want the markdown templates to be externalized into a `templates` folder with files with `.tpl.md` extension.
