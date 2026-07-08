# AI instruction 25

## Image croping

Replace completely the croping of images by storing x and y position and use it to display the images in the frontend.

In the pdf I don't want the image to be cropped.

Remove completely `REACT_APP_IMAGE_SIZE` and replace it with a `CWCLOCK_MAX_IMAGE_SIZE` in bytes environment variable on the backend side with a error 400 with `i18n_code` if it's too big (set it to 2MB by default).

Add also a `REACT_APP_MAX_IMAGE_SIZE` on the frontend to reject before calling the API.

Put it like this in the `.env.react`: `MAX_IMAGE_SIZE=${CWCLOCK_MAX_IMAGE_SIZE}` and defined it to 2MB in the `compute-env.sh`.

If it's not parsable as a number apply 2MB as default value.

## Delete buttons

For client and project add delete buttons in the management screens, add the webservices if those are missing (check if the user is a `superadmin`, or `owner` or `admin` of the organization).

## RBAC

For each action checking if the user is `owner` or `admin`, check also if it's a global `superadmin` (who should also be allowed to perform the action).
