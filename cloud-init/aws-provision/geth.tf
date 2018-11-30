resource "null_resource" "provision_ebs" {
  connection {
    host        = "${aws_instance.geth_instance.public_dns}"
    user        = "ubuntu"
    private_key = "${file(var.ssh_private_key_path)}"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo mkfs -t ext4 /dev/xvdg",
      "sudo mkdir /datadrive",
      "sudo mount /dev/xvdg /datadrive",
      "sudo echo /dev/xvdg /datadrive ext4 defaults,nofail 0 2 >> /etc/fstab",
      "sudo chown -R `whoami` /datadrive",
    ]
  }
}

resource "null_resource" "provision_nginx" {
  connection {
    host        = "${aws_instance.geth_instance.public_dns}"
    user        = "ubuntu"
    private_key = "${file(var.ssh_private_key_path)}"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo apt-get update",
      "sudo apt-get -y install nginx apache2-utils",
      "sudo htpasswd -bc /etc/nginx/.htpasswd ${random_string.username.result} ${random_string.password.result}",
    ]
  }

  provisioner "file" {
    source      = "${var.nginx_conf}"
    destination = "./${var.nginx_conf}"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo rm /etc/nginx/sites-enabled/* || true",
      "sudo mv ./${var.nginx_conf} /etc/nginx/sites-enabled/",
      "sudo service nginx reload",
    ]
  }
}

resource "null_resource" "provision_docker" {
  depends_on = ["null_resource.provision_nginx"]

  connection {
    host        = "${aws_instance.geth_instance.public_dns}"
    user        = "ubuntu"
    private_key = "${file(var.ssh_private_key_path)}"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo apt-get -y remove docker docker-engine docker.io",
      "sudo apt-get update",
      "sudo apt-get -y install apt-transport-https ca-certificates curl software-properties-common zip",
      "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -",
      "sudo add-apt-repository \"deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable\"",
      "sudo apt-get update",
      "sudo apt-get -y install docker-ce",
    ]
  }
}

resource "null_resource" "provision_geth" {
  depends_on = ["null_resource.provision_ebs", "null_resource.provision_docker"]

  connection {
    host        = "${aws_instance.geth_instance.public_dns}"
    user        = "ubuntu"
    private_key = "${file(var.ssh_private_key_path)}"
  }

  provisioner "remote-exec" {
    inline = [
      "sudo docker stop geth || true",
      "sudo docker rm geth || true",
      "sudo docker pull ethereum/client-go:stable",
      <<EOF
        sudo docker run \
          --restart always \
          --name geth \
          -p 8546:8545 -p 30303:30303 \
          -v /datadrive:/datadrive \
          -d ethereum/client-go:stable --rpc --rpcvhosts "*" --rpcaddr "0.0.0.0" --datadir /datadrive --syncmode fast
      EOF
      ,
    ]
  }
}
