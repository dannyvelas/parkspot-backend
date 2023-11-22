# Solution Architecture

## Domain
* We use the root domain park-spot.co. This is the location of our landing page. The backend and the frontend are located at subdomains of this root domain.
* Our domains were purchased from Google Domains.
* Google Domains was recently purchased by Squarespace. So, these domains will be transferred to a Squarespace account sometime soon. It is not sure when.
* All of our domains support HTTPS (443) traffic. We configure this with LetsEncrypt. This is discussed below.
#### Landing Page
* park-spot.co points to our landing page.
* If one were to go to www.park-spot.co, it would re-direct to park-spot.co
#### Front End
* app.park-spot.co points to our front-end application
* If one were to go to parkspotapp.com or www.parkspotapp.com it would redirect to app.park-spot.co
* If one were to go to lasvistasguestparkingpasses.com OR www.lasvistasguestparkingpasses.com, it would also redirect to app.park-spot.co
* If one were to go to www.app.park-spot.co, the browser will show you a connection error
#### Backend
* api.park-spot.co points to our back-end application
* if one were to go to www.api.park-spot.co, the browser will show you a connection error
* there are no redirects to api.park-spot.co
#### Chart
| Domain                                   | Points To                      | Cost                | Renews           |
|------------------------------------------+--------------------------------+---------------------+------------------|
| park-spot.co                             | landing page                   | $30.00              | June 01, 2024    |
| app.park-spot.co                         | frontend                       |                     |                  |
| api.park-spot.co                         | backend                        |                     |                  |
| parkspotapp.com                          | (redirect to) park-spot.co     | $12.00              | June 02, 2024    |
| www.parkspotapp.com                      | (redirect to) park-spot.co     |                     |                  |
| lasvistasguestparkingpasses.com          | (redirect to) app.park-spot.co | $12.00              | June 02, 2024    |
| www.lasvistasguestparkingpasses.com      | (redirect to) app.park-spot.co |                     |                  |

## Server
* Our server is hosted by [Vultr](https://www.vultr.com/).
* Vultr is a VPS hosting service (IAAS)
* We pay $6 monthly, paid monthly for one VPS with 1 vCPU, 1GB ram, and 25GB of SSD
    * $5 are server costs
    * $1 is for auto backups
* The VPS is using Debian 11 x64
* There is another VPS that is would be $3.50, with half the ram and 10GB of SSD, but it is not available in Miami

## Deployment
* To deploy the frontend, the backend, and the landing page, we use [dokku](https://dokku.com/).
* Dokku is a deployment tool that you can install on your server

### Setting up a dokku app for the first time
* Once you install it on your server, you can register a new application with dokku.
* Once you register a new application, you can make dokku become aware of the code base of the application by adding a new remote origin to the git of your application.
* Doing this will be very similar to adding a new git server to your application
* You can do this by issuing the remote add command at the root of your repository. The syntax for this is `git remote add dokku dokku@<root-domain-of-server>:<dokku-app-name>`
* For example, let's suppose we want to connect the landing page repo to the dokku remote origin of our server
    * The name of our landing page application in dokku is `parkspot-landing`
    * The root domain of our server is `park-spot.co`
    * The command for this would be: `git remote add dokku dokku@park-spot.co:parkspot-landing`
* After you make dokku a remote origin of your git repo, you can deploy your application to dokku.
* You can deploy to dokku by doing a git push that points to your dokku remote origin. This is the command: `git push dokku:main`
* Every time you have a new version of your local `main` branch, and you would like to push it to dokku, you can issue that command.
* [This is a very useful tutorial that I used](https://shellbear.me/blog/go-dokku-deployment) that has more detailed instructions on deploying an application
* Note: If you clone a repository on a new computer, and you want to push to dokku, you will have to re-associate the repo with dokku by running the `git remote add dokku dokku@<root-domain-of-server>:<dokku-app-name>` command again.

### Internals
* For each application, Dokku creates a docker container, it will build your application and run it within the container.
* On Dokku, you configure the root domain of an application will respond to.
    * Suppose I have a server with an IP address that is pointed to by `example.com`.
    * Suppose that on this server, I am using dokku.
    * Suppose that on dokku I registered an application which is an HTTP server called `mybackendapp` and I would like to deploy it to `backend.example.com`.
    * In Dokku, I can issue this command: `dokku domains:add mybackendapp backend.example.com`.
    * This will tell Dokku that `mybackendapp` should be deployed to the domain `backend.example.com`.
* Dokku uses NGINX as a proxy to route requests from the VPS to the docker container in which an app is running.
* If I were to run this command: `dokku domains:add mybackendapp backend.example.com`, Dokku will configure the NGINX configuration of the VPS to redirect all requests with the subdomain `backend`, to the Docker container running `mybackendapp`.
* It will work successfully as long as your application is listening to the IP address `0.0.0.0` inside of the docker container.

### Certificate Generation via LetsEncrypt
* Dokku allows you to create a certificate for your application with their LetsEncrypt plugin
* I'm not sure why but when I run the command to add a domain to an application, it will automatically attempt to create a certificate for that domain so that my domain has support for HTTPS.
* When I do this, the command will succeed in adding a domain to my application, but it will fail to create the certificate.
* When dokku tries to create the certificate it tries to hit the domain. If the domain does not resolve, the certificate will not be created.
* The certificate creation fails because the domain does not resolve. The domain does not resolve because I am in the process of adding the domain. The domain can't resolve if it has not yet finished being added.
* Once the command to add a domain returns, it gives two outputs. It says the domain creation worked but then says the certificate creation failed with an error like this: `acme: error: 403 :: urn:ietf:params:acme:error:unauthorized`.
* At this point, I re-run the command to add a domain and the certificate creation will work.

## Database hosting
* We also use dokku for this
* Dokku allows you to create a instance of a database as a "plugin". 
* For example, [this page](https://dokku.com/docs/deployment/application-deployment/#create-the-backing-services) has instructions on how to create a PostgreSQL instance and link it to an application on Dokku.
* This is the approach we take to host our PostgreSQL database.

## Landing Page
* Our landing page is what you see when you go to park-spot.co.
* It is exclusively static content: it is a single HTML file.
* It was deployed using [these instructions](https://johnfraney.ca/blog/build-deploy-static-site-dokku/).
* There is no caching layer. We do not use a CDN.
* Sin* The static content is on a docker container and it receives a requests forwarded from NGINX based on the route

## The frontend

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
