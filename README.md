# dmount - mount volumes on sibling container


* "-v" maps host files to a container
* "--volumes-from" maps a data volume to a container

What if you don't know which one to use?  `dmnt` fixes this in a generic way.

docker --rm -it \
  $(dmnt /usr/local/work/project) \
  container  ls -l

For host-to-container it will expand to:
```
docker --rm -it \
  -v /usr/local/work/project:/usr/local/work/project \
  container ls -l
```

for data volume to container it will expand to:
```
docker --rm -it \
  -v 7878dedbeaef37627838276386283648234:/usr/local/work \
  container ls -l
```

(or whatever the volume name is)


## The Problem

### Good: host to container

The standard case is mounting data from the host machine (that's running the docker
daemon) into the container

```bash
docker --rm -it \
  -v /Users/nickg/work:/usr/local/work \ 
  -w /go/src \
  my-dev-env ls -l
```

### Good: data volume to container

Now let's say I don't want to mess up my laptop (or server) with a bunch of
work stuff, and I so I want to keep everything in a [data volume]().

```bash
docker create -v /usr/local/work --name workspace alpine /bin/true
```

and now mounting this:

```
docker --rm -it \
 --volumes-from workspace \
 -w /usr/local/work \
 my-dev-env ls -l
```

Once I'm inside the container, everything I do in /usr/local/work will
be saved but the host is left untouched.  Basically I never need to use
`homebrew` or anything else.  My dev environment is 100% in docker.

### Good: Using Docker with in Docker

Now your development environment is in docker.   Since you have now completely 
drank the k00l-aid, you'll want to run or build containers inside docker.

This is not "docker in docker", but generated or run new contains _in
parallel_ to the container you have. [jpetazzo calls this a sibling container](
(http://jpetazzo.github.io/2015/09/03/do-not-use-docker-in-docker-for-ci/)
It's very nice. 

1. mount `/var/run/docker.sock`
2. add docker (the client) into your container

```sh
docker --rm -it \
 -v /var/run/docker.sock:/var/run/docker.sock \
 --volumes-from workspace \
 -w /usr/local/work \
  my-dev-env ls -l
```

### Ugly: Sharing data volumes with sibling containers


You'd like to do this

```sh
docker --rm -it \
   -v /usr/local/work/newproject:/usr/local/newproject \
   -w /usr/local/work/newproject \
   sibling-container ls -l

NOTHING!!
```

With the "pass the /var/run/docker.sock" trick, docker on the host
doesn't know it's being called from a container, and so
everything is relative to the host, _not the container_.

And `/usr/local/work/newproject` is likely nothing, or not what you want.

Instead for sibling containers, you need to redo the `--volumes-from` flags:

```sh
docker --rm -it \
   --volumes-from workspace \
   -w /usr/local/work/newproject \
   sibling-container ls -l

# hello files
```

So this works great _if you already know this_.  But what if you don't know
if `/usr/local/work/newproject` is mounted on the host or is a data volume?

In general should you use:

```
-v /usr/local/work/newproject:/usr/local/newproject
```

or

```
--volumes-from workspace
```

This is where `dmnt` comes in.  It will inspect the given directory, and see
if who owns it, and determines the correct command line to use.

The algorihm is the following:

* Can we connect to `/var/run/docker.sock` ? No use `-v /directory:/directory`
* run `docker inspect ${HOSTNAME}` (more or less) and get containers mount
  points
* If the directory is not exactly or contained in any of mount's destinations:
  use `-v /directory:/directory`
* If the mount doesn't have `Name` entry: use `-v /directory:/directory`
* Use `-v name:destination`

Enjoy!

