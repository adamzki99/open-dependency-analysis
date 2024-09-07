import hashlib
import tkinter as tk

from tqdm import tqdm
from pyvis.network import Network

class Edge:

    def __init__(self, a, b, w):
        self.node_a = a
        self.node_b = b
        self.weight = w

class Node:

    def __init__(self, v, l, c, s):
        self.value = v
        self.label = l
        self.color = c
        self.size = s



def get_screen_resolution():
    root = tk.Tk()
    root.withdraw()  # Hide the main window
    width = root.winfo_screenwidth()
    height = root.winfo_screenheight()
    return width, height

def string_to_color(input_string):
    # Hash the input string using SHA-256
    hash_object = hashlib.sha256(input_string.encode('utf-8'))
    hex_digest = hash_object.hexdigest()
    
    # Take the first 6 characters of the hash to form the hex color code
    hex_color = '#' + hex_digest[:6]
    
    return hex_color

def update_node_size(nodes, dynamic_node_size):


    print('Updating node sizes')

    if dynamic_node_size == True:
        pbar = tqdm(total=len(nodes)*2) # times 2 because of the two loops

        largest_node_size = -1

        # first get the size of the biggest node
        for node in nodes:
            pbar.update(1)
            if largest_node_size < node.size:
                largest_node_size = node.size

        for node in nodes:
            pbar.update(1)        
            node.size = (node.size / largest_node_size) * 20

        pbar.close()
    else:
        pbar = tqdm(total=len(nodes)*2)
        
        for node in nodes:
            pbar.update(1)        
            node.size = 10
        
        pbar.close()

def create_real_network_view(json_data, weight_threshold, dynamic_node_size, dynamic_edge_size):

    network_width, network_height = get_screen_resolution()

    network_width = round(network_width * 0.8, 0)
    network_height = round(network_height * 0.8, 0)
    

    net = Network(network_height, network_width)
    net.set_options('''
const options = {
  "nodes": {
    "borderWidth": 2,
    "borderWidthSelected": 4,
    "font": {
      "color": "#6c6c6c"
    },
    "shape": "circle"
  },
  "edges": {
    "color": {
      "color": "#eaeaea",
      "inherit": false
    },
    "smooth": {
      "type": "continuous"
    },
    "width": 0.1
  },
  "interaction": {
    "hover": true,
    "multiselect": true
  },
  "physics": {
    "forceAtlas2Based": {
      "springLength": 100,
      "avoidOverlap": 0.66
    },
    "minVelocity": 0.75,
    "solver": "forceAtlas2Based"
  }
}
''')

    nodes, edges = create_fully_connected_graph(json_data, weight_threshold)
    update_node_size(nodes, dynamic_node_size)

    for node in nodes:
        net.add_node(node.value, label=node.label, color=node.color, size=node.size)

    for edge in edges:
        try:
            if dynamic_edge_size == True:
                net.add_edge(edge.node_a, edge.node_b, value=edge.weight)
            else:
                net.add_edge(edge.node_a, edge.node_b)
        except:
            pass

    net.prep_notebook()

    net.show('da_network_view.html', notebook = False)

def create_fully_connected_graph(json_data, weight_threshold):
    
    # Initialize the list of edges
    edges = []
    
    # Set to track nodes that are part of valid edges
    connected_nodes = set()
    
    # Iterate over the upper triangle of the DataFrame (since it's a symmetric matrix)
    print(f'Including edges, based on threshold: {weight_threshold}')
    pbar = tqdm(total=len(json_data))
    for _, entry in enumerate(json_data):
        pbar.update(1)
        for edge_weight, node_b in enumerate(entry['References']):                

            # Include only edges with weight above the threshold
            if edge_weight >= weight_threshold:
                edges.append(Edge(entry['PackageName'], node_b, edge_weight))
                connected_nodes.update([entry['PackageName'], node_b])
    
    pbar.close()

    # Convert the set of connected nodes back to a list
    connected_nodes = list(connected_nodes)
    nodes = []

    print('Cleaning up nodes to be in network')
    pbar = tqdm(total=len(json_data))
    for entry in json_data:
        pbar.update(1)
        if entry['PackageName'] in connected_nodes:

            nodes.append(Node(entry['PackageName'], entry['PackageName'], string_to_color(entry['DirectoryName']), entry['NumberOfLines']))
    pbar.close()

    return nodes, edges
