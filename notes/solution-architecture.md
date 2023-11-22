# Solution Architecture

## Tiers 

* Servers hosted by [Vultr](https://www.vultr.com/).
    * Vultr is a VPS hosting service (IAAS)
    * We pay $6 monthly, paid monthly for one VPS with 1 vCPU, 1GB ram, and 25GB of SSD
        * $5 are server costs
        * $1 is for auto backups
    * The VPS is using Debian 11 x64
    * There is another VPS that is would be $3.50, with half the ram and 10GB of SSD, but it is not available in Miami
* To deploy the frontend, the backend, and the landing page, we use [dokku](https://dokku.com/).
    * Dokku is a deployment tool that you can install on your server
    * Setting up a dokku app for the first time
        * Once you install it on your server, you can register a new application with dokku.
        * Once you register a new application, you can deploy it to dokku by adding a new remote origin to the git of your application.
        * Doing this will be very similar to adding a new git server to your application
        * You can do this by issuing the following command at the root of your repository: `git remote add dokku dokku@park-spot.co:parkspot-landing`
        * Doing so will allow you to deploy your application to your server.
        * If you clone a repository on a new computer, you will have to re-associate it with dokku by running the `git remote add dokku dokku@park-spot.co:parkspot-landing` command again.
        * After this, you can push your application to dokku for the first time by running `git push dokku:main`
        * Every time you have a new version of your local `main` branch, and you would like to push it to dokku, you can issue the same command: `git push dokku:main`
    * Internals
        * For each application, Dokku creates a docker container, it will build your application and run it within the container.
        * On Dokku, you configure the root domain of an application will respond to.
            * Suppose I have a server with an IP address that is pointed to by `example.com`.
            * Suppose that on this server, I am using Dokku.
            * Suppose that on Dokku I registered an application which is an HTTP server called `mybackendapp` and I would like to deploy it to `backend.example.com`.
            * In Dokku, I can issue this command: `dokku domains:add mybackendapp backend.example.com`.
            * This will tell Dokku that `mybackendapp` should be deployed to the domain `backend.example.com`.
        * Dokku uses NGINX as a proxy to route requests from the VPS to the docker container in which an app is running.
        * If I were to run this command: `dokku domains:add mybackendapp backend.example.com`, Dokku will configure the NGINX configuration of the VPS to redirect all requests with the subdomain `backend`, to the Docker container running `mybackendapp`.
        * It will work successfully as long as your application is listening to the ip address `0.0.0.0` inside of the docker container.
* For database hosting, we also use dokku.
    * Dokku allows you to create a instance of a database as a "plugin". 
    * For example, [this page](https://dokku.com/docs/deployment/application-deployment/#create-the-backing-services) has instructions on how to create a PostgreSQL instance and link it to an application on Dokku.
    * This is the approach we take to host our PostgreSQL database.

### Landing Page
* Our landing page is what you see when you go to park-spot.co.
* It is exclusively static content: it is a single HTML file.
* It was deployed using [these instructions](https://johnfraney.ca/blog/build-deploy-static-site-dokku/).
* There is no caching layer. We do not use a CDN.
* Sin* The static content is on a docker container and it receives a requests forwarded from NGINX based on the route

### The frontend

The frontend is a SvelteKit application. 

When a client opens the website for the first time, the content that they will see will be server-side generated. When the client navigates to new pages on the website thereafter, the content of those pages will be generated on the client-side.

The website is divided into two:
* Static content: Landing page
    * Deployed separately to park-spot.co
    * No caching layer (this can be improved by using a CDN)
    * The static content is on a docker container and it receives a requests forwarded from NGINX based on the route
* Dynamic content:
    * Pages
        * The login page
        * every page after logging in
        * the forgot password page
    * The dynamic content is on a docker container and it receives requests forwarded from NGINX based on the route

### Backend

* The backend is a Go application.
* It uses a net/http application server.
* So, when you compile the backend, it will be one binary

### Database

* Right now, the database is abstracted by a "storage" layer which is a set of functions that take Go datatypes as arguments and return Go datatypes. These functions interface with the database directly.
* In the future, it might be worth considering to use an ORM instead of writing the logic of the storage layer: https://gorm.io/docs/.
