import sys
import yaml

def generate_compose(output_file, clients):
    with open(output_file) as f:
        compose = yaml.safe_load(f)
    
    for i in range(1, clients + 1):
        compose["services"][f"client{i}"] = {
            "container_name": f"client{i}",
            "image": "client:latest",
            "entrypoint": "/client",
            "environment": ["CLI_ID=1", "CLI_LOG_LEVEL=DEBUG"],
            "networks": ["testing_net"],
            "depends_on": ["server"],
        }

    # Save the new compose file
    with open(output_file, "w") as output:
        yaml_str = yaml.dump(compose, default_flow_style=False, sort_keys=False)
        # Add blank line between clients
        yaml_str = yaml_str.replace("  client", "\n  client")
        output.write(yaml_str)


if __name__ == '__main__':
    if len(sys.argv) >= 3:
        output_file = sys.argv[1]
        clients = sys.argv[2]
        generate_compose(output_file, int(clients))
    else:
        print(f"Missing arguments")


