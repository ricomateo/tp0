import sys
import yaml

def generate_compose(base_file, output_file, clients):
    with open(base_file) as f:
        compose = yaml.safe_load(f)
    
    compose["services"]["server"]["volumes"] = [{
        "type": "bind",
        "source": "./server/config.ini",
        "target": "/config.ini"
    }]
    # Empty up the server environment
    del compose["services"]["server"]["environment"]
    
    # Collect the clients to delete
    clients_to_delete = []
    for service in compose["services"].keys():
        if service.startswith("client"):
            clients_to_delete.append(service)
    
    # Delete the clients
    for client in clients_to_delete:
        del compose["services"][client]

    # Bet data (works with up to 5 clients)
    names = ["Juan", "Maria", "Nicolas", "Juana", "Santiago"]
    last_name = "Perez" # Same last name for all clients, for simplicty
    document = 40_000_000
    birthday = "2000-01-01"
    number = 100

    for i in range(1, clients + 1):
        compose["services"][f"client{i}"] = {
            "container_name": f"client{i}",
            "image": "client:latest",
            "entrypoint": "/client",
            "environment": [
                "CLI_ID=1",
                f"AGENCIA={i}",
                f"NOMBRE={names[i - 1]}",
                f"APELLIDO={last_name}",
                f"DOCUMENTO={document + i}",
                f"NACIMIENTO={birthday[:-1] + str(i)}",
                f"NUMERO={number + i}",
            ],
            "networks": ["testing_net"],
            "depends_on": ["server"],
            "volumes": [{
                "type": "bind",
                "source": "./client/config.yaml",
                "target": "/config.yaml"
            }]
        }

    # Save the new compose file
    with open(output_file, "w") as output:
        yaml_str = yaml.dump(compose, default_flow_style=False, sort_keys=False)
        # Add blank line between clients
        yaml_str = yaml_str.replace("  client", "\n  client")
        output.write(yaml_str)


if __name__ == '__main__':
    if len(sys.argv) >= 4:
        base_file = sys.argv[1]
        output_file = sys.argv[2]
        clients = sys.argv[3]
        generate_compose(base_file, output_file, int(clients))
    else:
        print(f"Missing arguments")


