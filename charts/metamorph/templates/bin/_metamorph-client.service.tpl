[Unit]
Description=MetaMorph Boot Actions Service
ConditionPathExists=!/var/install/metamorph-bootactions.lock

[Service]
Environment="HOME=/root"
Type=oneshot
ExecStartPre=-/bin/mkdir -p /var/install
ExecStart=/bin/bash /root/init.sh &

[Install]
WantedBy=multi-user.target
