maintainer        "Erik St. Martin"
maintainer_email  "alakriti@gmail.com"
license           "Apache 2.0"
description       "Installs go"
version           "1.0.0"

recipe "go", "Installs go"

%w{ debian ubuntu }.each do |os|
  supports os
end
