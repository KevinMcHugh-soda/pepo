This is a project to build a simple web application that supports tracking performance of direct reports.

The project is implemented in Golang.

It uses a RESTful API to communicate with the frontend. The API is defined using OpenAPI. We generate code for the API using https://ogen.dev/

We store the data in a PostgreSQL database. We generate code to interact with the database using https://sqlc.dev/

The database schema is managed using https://github.com/amacneil/dbmate

The core database model is a Person, who has a name. The primary key is an xid: https://github.com/rs/xid

The front-end will be implemented using HTMX: https://htmx.org/ and Tailwind CSS: https://tailwindcss.com/
