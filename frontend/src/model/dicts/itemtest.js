const url = 'https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/default/v1/items.json';

// Function to fetch and process the JSON data
async function fetchData() {
    try {
        const response = await fetch(url);
        if (!response.ok) {
            throw new Error('Network response was not ok ' + response.statusText);
        }
        const data = await response.json();

        const s = '/lol-game-data/assets/ASSETS/Items/Icons2D/'
        const result = data.reduce((acc, item) => {
            acc[item.id] = item.iconPath.substring(s.length);
            return acc;
        }, {});
        console.log(result);
    } catch (error) {
        console.error('There has been a problem with your fetch operation:', error);
    }
}

// Call the function to fetch and process the data
fetchData();