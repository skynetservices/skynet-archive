maintainer        "Erik St. Martin"
maintainer_email  "alakriti@gmail.com"
license           "Apache 2.0"
description       "Sets up example skynet services"
version           "1.0.0"

recipe "skynet-example", "Sets up example skynet services"

%w{ debian ubuntu }.each do |os|
  supports os
end
