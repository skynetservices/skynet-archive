maintainer        "Erik St. Martin"
maintainer_email  "alakriti@gmail.com"
license           "Apache 2.0"
description       "Install bzr revision control"
version           "1.0.0"
recipe            "bzr", "Installs bzr"

%w{ fedora redhat centos ubuntu debian }.each do |os|
  supports os
end
