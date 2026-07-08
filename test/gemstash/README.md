This exists because the dummy project entrypoint installs rails. We have a
caching feature that prevents the same dependencies being downloaded again
on a fork, however for the purposes of development and testing we want to
exercise both the caching and the non caching code paths. To avoid excessive
calls to rubygems we need a pull through cache
