#!/bin/bash

red='\033[0;31m'
green='\033[0;32m'
yellow='\033[0;33m'
plain='\033[0m'

#Add some basic function here
function LOGD() {
    echo -e "${yellow}[DEG] $* ${plain}"
}

function LOGE() {
    echo -e "${red}[ERR] $* ${plain}"
}

function LOGI() {
    echo -e "${green}[INF] $* ${plain}"
}
# check root
[[ $EUID -ne 0 ]] && LOGE "ERROR: You must be root to run this script! \n" && exit 1

# Check OS and set release variable
if [[ -f /etc/os-release ]]; then
    source /etc/os-release
    release=$ID
elif [[ -f /usr/lib/os-release ]]; then
    source /usr/lib/os-release
    release=$ID
else
    echo "Failed to check the system OS, please contact the author!" >&2
    exit 1
fi

echo "The OS release is: $release"


os_version=""
os_version=$(grep -i version_id /etc/os-release | cut -d \" -f2 | cut -d . -f1)

if [[ "${release}" == "centos" ]]; then
    if [[ ${os_version} -lt 8 ]]; then
        echo -e "${red} Please use CentOS 8 or higher ${plain}\n" && exit 1
    fi
elif [[ "${release}" ==  "ubuntu" ]]; then
    if [[ ${os_version} -lt 20 ]]; then
        echo -e "${red}please use Ubuntu 20 or higher version! ${plain}\n" && exit 1
    fi

elif [[ "${release}" == "fedora" ]]; then
    if [[ ${os_version} -lt 36 ]]; then
        echo -e "${red}please use Fedora 36 or higher version! ${plain}\n" && exit 1
    fi

elif [[ "${release}" == "debian" ]]; then
    if [[ ${os_version} -lt 10 ]]; then
        echo -e "${red} Please use Debian 10 or higher ${plain}\n" && exit 1
    fi
fi

confirm() {
    if [[ $# > 1 ]]; then
        echo && read -p "$1 [Default$2]: " temp
        if [[ x"${temp}" == x"" ]]; then
            temp=$2
        fi
    else
        read -p "$1 [y/n]: " temp
    fi
    if [[ x"${temp}" == x"y" || x"${temp}" == x"Y" ]]; then
        return 0
    else
        return 1
    fi
}

confirm_restart() {
    confirm "Restart $1" "y"
    if [[ $? == 0 ]]; then
        restart "$1"
    else
        show_menu
    fi
}

before_show_menu() {
    echo && echo -n -e "${yellow}Press enter to return to the main menu: ${plain}" && read temp
    show_menu
}

install() {
    bash <(curl -Ls https://raw.githubusercontent.com/Raha-Project/raha/main/install.sh)
    if [[ $? == 0 ]]; then
        if [[ $# == 0 ]]; then
            start "raha-xray"
        else
            start "raha-xray" 0
        fi
    fi
}

update() {
    confirm "This function will forcefully reinstall the latest version, and the data will not be lost. Do you want to continue?" "n"
    if [[ $? != 0 ]]; then
        LOGE "Cancelled"
        if [[ $# == 0 ]]; then
            before_show_menu
        fi
        return 0
    fi
    bash <(curl -Ls https://raw.githubusercontent.com/Raha-Project/raha/main/install.sh)
    if [[ $? == 0 ]]; then
        LOGI "Update is complete, app has automatically restarted"
        exit 0
    fi
}

uninstall() {
    confirm "Are you sure you want to uninstall the app? xray will also uninstalled!" "n"
    if [[ $? != 0 ]]; then
        if [[ $# == 0 ]]; then
            show_menu
        fi
        return 0
    fi
    systemctl stop raha-xray
    systemctl disable raha-xray
    rm /etc/systemd/system/raha-xray.service -f
    systemctl stop xray
    systemctl disable xray
    rm /etc/systemd/system/xray.service -f

    systemctl daemon-reload
    systemctl reset-failed
    rm /usr/local/raha-xray/ -rf

    echo ""
    echo -e "Uninstall succeed! If you want to remove this script, then after exiting the script run ${green}rm /usr/bin/raha-xray -f${plain} to delete it."
    echo ""

    if [[ $# == 0 ]]; then
        before_show_menu
    fi
}

reset_config() {
    local config_file="/usr/local/raha-xray/raha-xray.json"

    # Check if the file exists
    if [ ! -f "$config_file" ]; then
        echo "Error: Configuration of raha-xray is not in place!"
        return
    else
        # Load JSON content into a variable
        local json_content
        config_json=$(<"$config_file")
    fi

    local options=("listen" "domain" "port" "certFile" "keyFile" "basePath" "timeLocation" "dbType" "dbAddr")

    # Display a menu for editing options
    while true; do
        echo "Current Configuration Menu:"
        for i in "${!options[@]}"; do
            defaultValue=$(echo "$config_json" | jq ".${options[$i]}")
            echo "$i) Edit ${options[$i]} (${defaultValue})"
        done
        echo "9) Save & Exit"
        echo ""

        read -rp "Enter the number of the option to edit (or '9' to save and exit): " choice

        case $choice in
            0) read -rp "Enter Listen value: " listen; config_json=$(echo "$config_json" | jq ".listen=\"$listen\"") ;;
            1) read -rp "Enter Domain value: " domain; config_json=$(echo "$config_json" | jq ".domain=\"$domain\"") ;;
            2) read -rp "Enter Port value: " port; config_json=$(echo "$config_json" | jq ".port=$port") ;;
            3) read -rp "Enter CertFile value: " certFile; config_json=$(echo "$config_json" | jq ".certFile=\"$certFile\"") ;;
            4) read -rp "Enter KeyFile value: " keyFile; config_json=$(echo "$config_json" | jq ".keyFile=\"$keyFile\"") ;;
            5) read -rp "Enter BasePath value: " basePath; config_json=$(echo "$config_json" | jq ".basePath=\"$basePath\"") ;;
            6) read -rp "Enter TimeLocation value: " timeLocation; config_json=$(echo "$config_json" | jq ".timeLocation=\"$timeLocation\"") ;;
            7) read -rp "Enter DbType value: " dbType; config_json=$(echo "$config_json" | jq ".dbType=\"$dbType\"") ;;
            8) read -rp "Enter DbAddr value: " dbAddr; config_json=$(echo "$config_json" | jq ".dbAddr=\"$dbAddr\"") ;;
            9) break ;; # Exit
            *) echo "Invalid option" ;;
        esac
    done

    # Save the updated JSON
    echo "$config_json" > "$config_file"

    echo "Configuration updated successfully."
    if [[ $# == 0 ]]; then
        confirm_restart "raha-xray"
    fi
}

show_token(){
    cd /usr/local/raha-xray/
    if [[ $# == 0 ]]; then
        read -p "Enter token ID: " tokenId
        ./rahaXray token -id ${tokenId}
    else
        ./rahaXray token -list
    fi
    before_show_menu
}

add_token(){
    cd /usr/local/raha-xray/
    ./rahaXray token add
    if [[ $? == 0 ]]; then
        LOGI "New token added Successfully"
    else
        LOGE "Failed to add new token"
    fi
    before_show_menu
}

del_token(){
    read -p "Enter token ID to delete: " tokenId
    cd /usr/local/raha-xray/
    ./rahaXray token -del ${tokenId}
    if [[ $? == 0 ]]; then
        LOGI "Token $tokenId deleted Successfully"
    else
        LOGE "Failed to delete token $tokenId"
    fi
    before_show_menu
}

start() {
    check_status "$1"
    if [[ $? == 0 ]]; then
        echo ""
        LOGI "$1 is running, No need to start again, If you need to restart, please select restart"
    else
        systemctl start "$1"
        sleep 2
        check_status "$1"
        if [[ $? == 0 ]]; then
            LOGI "raha-xray Started Successfully"
        else
            LOGE "Failed to start $1, Probably because it takes longer than two seconds to start, Please check the log information later"
        fi
    fi

    if [[ $# == 1 ]]; then
        before_show_menu
    fi
}

stop() {
    check_status "$1"
    if [[ $? == 1 ]]; then
        echo ""
        LOGI "API is aleady stopped, No need to stop again!"
    else
        systemctl stop  "$1"
        sleep 2
        check_status  "$1"
        if [[ $? == 1 ]]; then
            LOGI "$1 stopped successfully"
        else
            LOGE "Failed to stop $1, Probably because the stop time exceeds two seconds, Please check the log information later"
        fi
    fi

    if [[ $# == 1 ]]; then
        before_show_menu
    fi
}

restart() {
    systemctl restart "$1"
    sleep 2
    check_status  "$1"
    if [[ $? == 0 ]]; then
        LOGI "$1 restarted successfully"
    else
        LOGE "Failed to restart $1, Probably because it takes longer than two seconds to start, Please check the log information later"
    fi
    if [[ $# == 1 ]]; then
        before_show_menu
    fi
}

status() {
    systemctl status  "$1" -l
    if [[ $# == 1 ]]; then
        before_show_menu
    fi
}

enable() {
    systemctl enable  "$1"
    if [[ $? == 0 ]]; then
        LOGI "Set $1 to boot automatically on startup successfully"
    else
        LOGE "Failed to set $1 Autostart"
    fi

    if [[ $# == 1 ]]; then
        before_show_menu
    fi
}

disable() {
    systemctl disable $1
    if [[ $? == 0 ]]; then
        LOGI "$1 autostart cancelled successfully"
    else
        LOGE "$1 Failed to cancel autostart"
    fi

    if [[ $# == 1 ]]; then
        before_show_menu
    fi
}

show_log() {
    journalctl -u "$1" -e --no-pager -f
    if [[ $# == 1 ]]; then
        before_show_menu
    fi
}

install_bbr() {
    # temporary workaround for installing bbr
    bash <(curl -L -s https://raw.githubusercontent.com/teddysun/across/master/bbr.sh)
    echo ""
    before_show_menu
}

# 0: running, 1: not running, 2: not installed
check_status() {
    if [[ ! -f "/etc/systemd/system/$1.service" ]]; then
        return 2
    fi
    temp=$(systemctl status "$1" | grep Active | awk '{print $3}' | cut -d "(" -f2 | cut -d ")" -f1)
    if [[ x"${temp}" == x"running" ]]; then
        return 0
    else
        return 1
    fi
}

check_enabled() {
    temp=$(systemctl is-enabled "$1")
    if [[ x"${temp}" == x"enabled" ]]; then
        return 0
    else
        return 1
    fi
}

check_uninstall() {
    check_status "raha-xray"
    if [[ $? != 2 ]]; then
        echo ""
        LOGE "Raha-xray is installed, Please do not reinstall"
        if [[ $# == 0 ]]; then
            before_show_menu
        fi
        return 1
    else
        return 0
    fi
}

check_install() {
    check_status "raha-xray"
    if [[ $? == 2 ]]; then
        echo ""
        LOGE "Please install raha-xray first"
        if [[ $# == 0 ]]; then
            before_show_menu
        fi
        return 1
    else
        return 0
    fi
}

show_status() {
    check_status "$1"
    case $? in
    0)
        echo -e "$1 state: ${green}Runing${plain}"
        show_enable_status "$1"
        ;;
    1)
        echo -e "$1 state: ${yellow}Not Running${plain}"
        show_enable_status "$1"
        ;;
    2)
        echo -e "$1 state: ${red}Not Installed${plain}"
        ;;
    esac
}

show_enable_status() {
    check_enabled "$1"
    if [[ $? == 0 ]]; then
        echo -e "Start automatically: ${green}Yes${plain}"
    else
        echo -e "Start automatically: ${red}No${plain}"
    fi
}

install_acme() {
    cd ~
    LOGI "install acme..."
    curl https://get.acme.sh | sh
    if [ $? -ne 0 ]; then
        LOGE "install acme failed"
        return 1
    else
        LOGI "install acme succeed"
    fi
    return 0
}

ssl_cert_issue_main() {
    echo -e "${green}\t1.${plain} Get SSL"
    echo -e "${green}\t2.${plain} Revoke"
    echo -e "${green}\t3.${plain} Force Renew"
    read -p "Choose an option: " choice
    case "$choice" in
        1) ssl_cert_issue ;;
        2) 
            local domain=""
            read -p "Please enter your domain name to revoke the certificate: " domain
            ~/.acme.sh/acme.sh --revoke -d ${domain}
            LOGI "Certificate revoked"
            ;;
        3)
            local domain=""
            read -p "Please enter your domain name to forcefully renew an SSL certificate: " domain
            ~/.acme.sh/acme.sh --renew -d ${domain} --force ;;
        *) echo "Invalid choice" ;;
    esac
}

ssl_cert_issue() {
    # check for acme.sh first
    if ! command -v ~/.acme.sh/acme.sh &>/dev/null; then
        echo "acme.sh could not be found. we will install it"
        install_acme
        if [ $? -ne 0 ]; then
            LOGE "install acme failed, please check logs"
            exit 1
        fi
    fi
    # install socat second
    case "${release}" in
        ubuntu|debian)
            apt update && apt install socat -y ;;
        centos)
            yum -y update && yum -y install socat ;;
        fedora)
            dnf -y update && dnf -y install socat ;;
        *)
            echo -e "${red}Unsupported operating system. Please check the script and install the necessary packages manually.${plain}\n"
            exit 1 ;;
    esac
    if [ $? -ne 0 ]; then
        LOGE "install socat failed, please check logs"
        exit 1
    else
        LOGI "install socat succeed..."
    fi

    # get the domain here,and we need verify it
    local domain=""
    read -p "Please enter your domain name:" domain
    LOGD "your domain is:${domain},check it..."
    # here we need to judge whether there exists cert already
    local currentCert=$(~/.acme.sh/acme.sh --list | tail -1 | awk '{print $1}')

    if [ ${currentCert} == ${domain} ]; then
        local certInfo=$(~/.acme.sh/acme.sh --list)
        LOGE "system already has certs here,can not issue again,current certs details:"
        LOGI "$certInfo"
        exit 1
    else
        LOGI "your domain is ready for issuing cert now..."
    fi

    # create a directory for install cert
    certPath="/root/cert/${domain}"
    if [ ! -d "$certPath" ]; then
        mkdir -p "$certPath"
    else
        rm -rf "$certPath"
        mkdir -p "$certPath"
    fi

    # get needed port here
    local WebPort=80
    read -p "please choose which port do you use,default will be 80 port:" WebPort
    if [[ ${WebPort} -gt 65535 || ${WebPort} -lt 1 ]]; then
        LOGE "your input ${WebPort} is invalid,will use default port"
    fi
    LOGI "will use port:${WebPort} to issue certs,please make sure this port is open..."
    # NOTE:This should be handled by user
    # open the port and kill the occupied progress
    ~/.acme.sh/acme.sh --set-default-ca --server letsencrypt
    ~/.acme.sh/acme.sh --issue -d ${domain} --standalone --httpport ${WebPort}
    if [ $? -ne 0 ]; then
        LOGE "issue certs failed,please check logs"
        rm -rf ~/.acme.sh/${domain}
        exit 1
    else
        LOGE "issue certs succeed,installing certs..."
    fi
    # install cert
    ~/.acme.sh/acme.sh --installcert -d ${domain} \
        --key-file /root/cert/${domain}/privkey.pem \
        --fullchain-file /root/cert/${domain}/fullchain.pem

    if [ $? -ne 0 ]; then
        LOGE "install certs failed,exit"
        rm -rf ~/.acme.sh/${domain}
        exit 1
    else
        LOGI "install certs succeed,enable auto renew..."
    fi

    ~/.acme.sh/acme.sh --upgrade --auto-upgrade
    if [ $? -ne 0 ]; then
        LOGE "auto renew failed, certs details:"
        ls -lah cert/*
        chmod 755 $certPath/*
        exit 1
    else
        LOGI "auto renew succeed, certs details:"
        ls -lah cert/*
        chmod 755 $certPath/*
    fi
}

ssl_cert_issue_CF() {
    echo -E ""
    LOGD "******Instructions for use******"
    LOGI "This Acme script requires the following data:"
    LOGI "1.Cloudflare Registered e-mail"
    LOGI "2.Cloudflare Global API Key"
    LOGI "3.The domain name that has been resolved dns to the current server by Cloudflare"
    LOGI "4.The script applies for a certificate. The default installation path is /root/cert "
    confirm "Confirmed?[y/n]" "y"
    if [ $? -eq 0 ]; then
        # check for acme.sh first
        if ! command -v ~/.acme.sh/acme.sh &>/dev/null; then
            echo "acme.sh could not be found. we will install it"
            install_acme
            if [ $? -ne 0 ]; then
                LOGE "install acme failed, please check logs"
                exit 1
            fi
        fi
        CF_Domain=""
        CF_GlobalKey=""
        CF_AccountEmail=""
        certPath=/root/cert
        if [ ! -d "$certPath" ]; then
            mkdir $certPath
        else
            rm -rf $certPath
            mkdir $certPath
        fi
        LOGD "Please set a domain name:"
        read -p "Input your domain here:" CF_Domain
        LOGD "Your domain name is set to:${CF_Domain}"
        LOGD "Please set the API key:"
        read -p "Input your key here:" CF_GlobalKey
        LOGD "Your API key is:${CF_GlobalKey}"
        LOGD "Please set up registered email:"
        read -p "Input your email here:" CF_AccountEmail
        LOGD "Your registered email address is:${CF_AccountEmail}"
        ~/.acme.sh/acme.sh --set-default-ca --server letsencrypt
        if [ $? -ne 0 ]; then
            LOGE "Default CA, Lets'Encrypt fail, script exiting..."
            exit 1
        fi
        export CF_Key="${CF_GlobalKey}"
        export CF_Email=${CF_AccountEmail}
        ~/.acme.sh/acme.sh --issue --dns dns_cf -d ${CF_Domain} -d *.${CF_Domain} --log
        if [ $? -ne 0 ]; then
            LOGE "Certificate issuance failed, script exiting..."
            exit 1
        else
            LOGI "Certificate issued Successfully, Installing..."
        fi
        ~/.acme.sh/acme.sh --installcert -d ${CF_Domain} -d *.${CF_Domain} --ca-file /root/cert/ca.cer \
        --cert-file /root/cert/${CF_Domain}.cer --key-file /root/cert/${CF_Domain}.key \
        --fullchain-file /root/cert/fullchain.cer
        if [ $? -ne 0 ]; then
            LOGE "Certificate installation failed, script exiting..."
            exit 1
        else
            LOGI "Certificate installed Successfully,Turning on automatic updates..."
        fi
        ~/.acme.sh/acme.sh --upgrade --auto-upgrade
        if [ $? -ne 0 ]; then
            LOGE "Auto update setup Failed, script exiting..."
            ls -lah cert
            chmod 755 $certPath
            exit 1
        else
            LOGI "The certificate is installed and auto-renewal is turned on, Specific information is as follows"
            ls -lah cert
            chmod 755 $certPath
        fi
    else
        show_menu
    fi
}

show_usage() {
    echo "raha-xray control menu usages: "
    echo "------------------------------------------"
    echo "raha-xray              - Enter     Admin menu"
    echo "raha-xray config       - Configure raha-xray"
    echo "raha-xray start        - Start     raha-xray"
    echo "raha-xray stop         - Stop      raha-xray"
    echo "raha-xray restart      - Restart   raha-xray"
    echo "raha-xray status       - Show      raha-xray status"
    echo "raha-xray enable       - Enable    raha-xray on system startup"
    echo "raha-xray disable      - Disable   raha-xray on system startup"
    echo "raha-xray log          - Check     raha-xray logs"
    echo "raha-xray update       - Update    raha-xray"
    echo "raha-xray install      - Install   raha-xray"
    echo "raha-xray uninstall    - Uninstall raha-xray"
    echo "------------------------------------------"
}

show_menu() {
    echo -e "
  ${green}raha-xray Management Script${plain}
  ${green}0.${plain} exit script
————————————————
  ${green}1.${plain} Install raha-xray
  ${green}2.${plain} Update raha-xray
  ${green}3.${plain} Uninstall raha-xray
————————————————
  ${green}4.${plain} Configure raha-xray
  ${green}5.${plain} Show all the Tokens
  ${green}6.${plain} Show a Token by ID
  ${green}7.${plain} Add a token
  ${green}8.${plain} Remove a token by ID
————————————————
  ${green}9.${plain} Start raha-xray
  ${green}10.${plain} Stop raha-xray
  ${green}11.${plain} Restart raha-xray
  ${green}12.${plain} Check raha-xray state
  ${green}13.${plain} Check raha-xray logs
  ${green}14.${plain} Set raha-xray Autostart
  ${green}15.${plain} Cancel raha-xray Autostart
————————————————
  ${green}16.${plain} Start xray
  ${green}17.${plain} Stop xray
  ${green}18.${plain} Restart xray
  ${green}19.${plain} Check xray state
  ${green}20.${plain} Check xray logs
  ${green}21.${plain} Set xray Autostart
  ${green}22.${plain} Cancel xray Autostart
————————————————
  ${green}23.${plain} 一A key installation bbr (latest kernel)
  ${green}24.${plain} 一SSL Certificate Management
  ${green}25.${plain} 一Cloudflare SSL Certificate
 "
    show_status "raha-xray"
    show_status "xray"
    echo && read -p "Please enter your selection [0-25]: " num

    case "${num}" in
    0)
        exit 0
        ;;
    1)
        check_uninstall && install
        ;;
    2)
        check_install && update
        ;;
    3)
        check_install && uninstall
        ;;
    4)
        check_install && reset_config
        ;;
    5)
        check_install && show_token all
        ;;
    6)
        check_install && show_token
        ;;
    7)
        check_install && add_token
        ;;
    8)
        check_install && del_token
        ;;
    9)
        check_install && start raha-xray
        ;;
    10)
        check_install && stop raha-xray
        ;;
    11)
        check_install && restart raha-xray
        ;;
    12)
        check_install && status raha-xray
        ;;
    13)
        check_install && show_log raha-xray
        ;;
    14)
        check_install && enable raha-xray
        ;;
    15)
        check_install && disable raha-xray
        ;;
    16)
        check_install && start xray
        ;;
    17)
        check_install && stop xray
        ;;
    18)
        check_install && restart xray
        ;;
    19)
        check_install && status xray
        ;;
    20)
        check_install && show_log xray
        ;;
    21)
        check_install && enable xray
        ;;
    22)
        check_install && disable xray
        ;;
    23)
        install_bbr
        ;;
    24)
        ssl_cert_issue_main
        ;;
    25)
        ssl_cert_issue_CF
        ;;
    *)
        LOGE "Please enter the correct number [0-16]"
        ;;
    esac
}

if [[ $# > 0 ]]; then
    case $1 in
    "start")
        check_install 0 && start raha-xray 0
        ;;
    "config")
        check_install 0 &&  reset_config 0
        ;;
    "stop")
        check_install 0 && stop raha-xray 0
        ;;
    "restart")
        check_install 0 && restart raha-xray 0
        ;;
    "status")
        check_install 0 && status raha-xray 0
        ;;
    "enable")
        check_install 0 && enable raha-xray 0
        ;;
    "disable")
        check_install 0 && disable raha-xray 0
        ;;
    "log")
        check_install 0 && logs raha-xray 0
        ;;
    "update")
        check_install 0 && update 0
        ;;
    "install")
        check_uninstall 0 && install 0
        ;;
    "uninstall")
        check_install 0 && uninstall 0
        ;;
    *) show_usage ;;
    esac
else
    show_menu
fi
