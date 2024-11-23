from diagrams import Cluster, Diagram
from diagrams.generic.device import Mobile
from diagrams.onprem.container import Docker
from diagrams.onprem.queue import RabbitMQ

with Diagram("Scalable chat app", show=False):
    # Clients (Mobile Devices)
    client1 = Mobile("Client 1")
    client2 = Mobile("Client 2")
    clients = [client1,
               client2]

    # WebSocket Servers (Using Generic Compute to represent servers)
    with Cluster("WebSocket Servers"):
        server1 = Docker("WS server 1")
        server2 = Docker("WS server 2")
        ws_servers = [server1, server2]

    # RabbitMQ in Fanout Mode
    with Cluster("RabbitMQ Fanout Exchange"):
        queue = RabbitMQ("message-queue")

    # Define the interactions
    client1 >> server1
    client2 >> server2

    for server in ws_servers:
        server - queue  # WebSocket Servers publish messages to RabbitMQ Exchange

    # queue >> ws_servers  # WebSocket Servers consume messages from Queue