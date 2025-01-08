async function createUser() {
    try {
        const name = document.getElementById('name')?.value.trim();
        const email = document.getElementById('email')?.value.trim();
        const password = document.getElementById('password')?.value.trim();
        const role = document.getElementById('role')?.value.trim().toLowerCase();
        if (!name) {
            alert('Name is required.');
            return;
        }
        if (name.length < 1) {
            alert('Name must be at least 1 characters long.');
            return;
        }
        if (!email) {
            alert('Email is required.');
            return;
        }
        if (!/^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$/i.test(email)) {
            alert('Invalid email format.');
            return;
        }
        if (!password) {
            alert('Password is required.');
            return;
        }
        if (password.length < 6) {
            alert('Password must be at least 6 characters long.');
            return;
        }
        const response = await fetch('/create', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, email, password, role }),
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(`Error creating user: ${error}`);
        }

        const result = await response.json();
        alert(`User created successfully: ${JSON.stringify(result)}`);
    } catch (err) {
        console.error('Error in createUser:', err);
        await reportClientError(err.message, 'createUser', null, null, err.stack || null);
        alert(`Failed to create user: ${err.message}`);
    }
}


let currentPage = 1;

async function getUsers(page = 1) {
    try {
        const response = await fetch(`/read?page=${page}`);
        if (!response.ok) throw new Error('Failed to fetch users.');

        const users = await response.json();
        let output = '<table border="1"><tr><th>ID</th><th>Name</th><th>Email</th><th>Password</th><th>Created At</th><th>Updated At</th></tr>';
        users.forEach(user => {
            output += `<tr>
                <td>${user.id}</td>
                <td>${user.name}</td>
                <td>${user.email}</td>
                <td>${user.password}</td>
                <td>${user.created_at}</td>
                <td>${user.updated_at}</td>
            </tr>`;
        });
        output += '</table>';

        output += `
            <button onclick="changePage(-1)">Previous</button>
            <button onclick="changePage(1)">Next</button>
        `;
        document.getElementById('output').innerHTML = output;
        currentPage = page;
    } catch (err) {
        console.error('Error in getUsers:', err);
        alert(`Failed to fetch users: ${err.message}`);
    }
}

function changePage(delta) {
    const nextPage = currentPage + delta;
    if (nextPage > 0) {
        getUsers(nextPage);
    }
}


async function updateUser() {
    try {
        const id = prompt('Enter User ID to update:');
        if (!id || isNaN(id) || parseInt(id) <= 0) throw new Error('Invalid User ID. Please enter a positive number.');

        const name = prompt('Enter new name (leave blank to keep current):');
        if (name && name.length < 3) throw new Error('Name must be at least 3 characters long.');

        const email = prompt('Enter new email (leave blank to keep current):');
        if (email && !/^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$/i.test(email)) throw new Error('Invalid email format.');

        const password = prompt('Enter new password (leave blank to keep current):');
        if (password && password.length < 6) throw new Error('Password must be at least 6 characters long.');


        

    const role = prompt('Enter new role (user/admin, leave blank to keep current):');
    if (role && (role !== "user" && role !== "admin")) {
        alert('Role must be either "user" or "admin".');
        return;
    }

    const response = await fetch('/update', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: parseInt(id), name, email, password, role }),
    });

        const result = await response.json();
        alert(`User updated successfully: ${JSON.stringify(result)}`);
    } catch (err) {
        console.error('Error in updateUser:', err);
        await reportClientError(err.message, 'updateUser', null, null, err.stack || null);
        alert(`Failed to update user: ${err.message}`);
    }
}

async function deleteUser() {
    try {
        const id = prompt('Enter User ID to delete:');
        if (!id || isNaN(id) || parseInt(id) <= 0) throw new Error('Invalid User ID. Please enter a positive number.');

        const response = await fetch('/delete', {
            method: 'DELETE',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ id: parseInt(id) }),
        });

        if (!response.ok) {
            const error = await response.text();
            throw new Error(`Error deleting user: ${error}`);
        }

        const result = await response.json();
        alert(`User deleted successfully: ${JSON.stringify(result)}`);
    } catch (err) {
        console.error('Error in deleteUser:', err);
        await reportClientError(err.message, 'deleteUser', null, null, err.stack || null);
        alert(`Failed to delete user: ${err.message}`);
    }
}

async function getUserByID() {
    try {
        const id = document.getElementById('userID').value;
        if (!id) throw new Error('Please enter a User ID.');

        const response = await fetch(`/readByID?id=${id}`);
        if (!response.ok) {
            const error = await response.text();
            throw new Error(`Error fetching user: ${error}`);
        }

        const user = await response.json();
        let output = `<table border="1"><tr><th>ID</th><th>Name</th><th>Email</th><th>Password</th><th>Role</th><th>Created At</th><th>Updated At</th></tr>`;
        output += `<tr>
            <td>${user.id}</td>
            <td>${user.name}</td>
            <td>${user.email}</td>
            <td>${user.password}</td>
            <td>${user.role}</td>
            <td>${user.created_at}</td>
            <td>${user.updated_at}</td>
        </tr>`;
        output += `</table>`;
        document.getElementById('output').innerHTML = output;
    } catch (err) {
        console.error('Error in getUserByID:', err);
        await reportClientError(err.message, 'getUserByID', null, null, err.stack || null);
        alert(`Failed to fetch user by ID: ${err.message}`);
    }
}
async function loadSampleData() {
    try {
        const sampleData = [
            { name: "Product 1", description: "Description of Product 1", price: 100, characteristics: "Feature A", date: "2025-01-01", image: "image1.jpg" },
            { name: "Product 2", description: "Description of Product 2", price: 200, characteristics: "Feature B", date: "2025-01-02", image: "image2.jpg" },
        ];

        for (const item of sampleData) {
            const response = await fetch('/create-product', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(item),
            });

            if (!response.ok) {
                const error = await response.text();
                throw new Error(`Error loading item: ${error}`);
            }
        }

        alert("Sample data successfully loaded!");
    } catch (err) {
        console.error('Error in loadSampleData:', err);
        alert(`Failed to load data: ${err.message}`);
    }
}
async function fetchAndDisplayProducts() {
    try {
        const response = await fetch('https://fakestoreapi.com/products');
        console.log(response.status)
        if (!response.ok) throw new Error('Failed to fetch products.');

        const products = await response.json();
        let output = '<table border="1"><tr><th>ID</th><th>Title</th><th>Price</th><th>Description</th><th>Category</th></tr>';
        products.forEach(product => {
            output += `<tr>
                <td>${product.id}</td>
                <td>${product.title}</td>
                <td>${product.price}</td>
                <td>${product.description}</td>
                <td>${product.category}</td>
            </tr>`;
        });
        output += '</table>';

        document.getElementById('products-output').innerHTML = output;
    } catch (err) {
        console.error('Error in fetchAndDisplayProducts:', err);
        alert(`Failed to fetch products: ${err.message}`);
    }
}
async function generateFakeUsers() {
    try {
        const fakeUsers = [];
        const numberOfUsers = 30; 

        for (let i = 1; i <= numberOfUsers; i++) {
            const user = {
                name: `User${i} Name`,
                email: `user${i}@example.com`,
                password: `password${i}`,
                role: 'user', 
            };
            fakeUsers.push(user);
        }

        for (const user of fakeUsers) {
            const response = await fetch('/create', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(user),
            });

            if (!response.ok) {
                const error = await response.text();
                throw new Error(`Failed to create user: ${error}`);
            }
        }

        alert(`${numberOfUsers} fake users generated successfully!`);
    } catch (err) {
        console.error('Error in generateFakeUsers:', err);
        alert(`Failed to generate users: ${err.message}`);
    }
}

async function filterUsers() {
    const name = document.getElementById('filterName').value.trim();
    const email = document.getElementById('filterEmail').value.trim();

    try {
        const params = new URLSearchParams();
        if (name) params.append("name", name);
        if (email) params.append("email", email);

        const response = await fetch(`/filter?${params.toString()}`);
        if (!response.ok) throw new Error('Failed to fetch filtered users.');

        const users = await response.json();
        let output = '<table border="1"><tr><th>ID</th><th>Name</th><th>Email</th><th>Created At</th><th>Updated At</th></tr>';
        users.forEach(user => {
            output += `<tr>
                <td>${user.id}</td>
                <td>${user.name}</td>
                <td>${user.email}</td>
                <td>${user.created_at}</td>
                <td>${user.updated_at}</td>
            </tr>`;
        });
        output += '</table>';
        document.getElementById('filterOutput').innerHTML = output;
    } catch (err) {
        console.error('Error in filterUsers:', err);
        alert(`Failed to apply filters: ${err.message}`);
    }
}

async function sortUsers() {
    const sortField = document.getElementById('sortField').value;
    const sortOrder = document.getElementById('sortOrder').value;

    try {
        const params = new URLSearchParams();
        params.append("field", sortField);
        params.append("order", sortOrder);

        const response = await fetch(`/sort?${params.toString()}`);
        if (!response.ok) throw new Error('Failed to fetch sorted users.');

        const users = await response.json();
        let output = '<table border="1"><tr><th>ID</th><th>Name</th><th>Email</th><th>Created At</th><th>Updated At</th></tr>';
        users.forEach(user => {
            output += `<tr>
                <td>${user.id}</td>
                <td>${user.name}</td>
                <td>${user.email}</td>
                <td>${user.created_at}</td>
                <td>${user.updated_at}</td>
            </tr>`;
        });
        output += '</table>';
        document.getElementById('sortOutput').innerHTML = output;
    } catch (err) {
        console.error('Error in sortUsers:', err);
        alert(`Failed to apply sort: ${err.message}`);
    }
}

async function reportClientError(errorMessage, source, line, column, stack) {
    const errorDetails = {
        message: errorMessage,
        source: source || 'N/A',
        line: line || 0,
        column: column || 0,
        stack: stack || 'N/A',
    };

    const logPayload = {
        action: 'clientError',
        status: 'error',
        details: errorDetails,
        time: new Date().toISOString(),
    };

    try {
        const response = await fetch('http://localhost:8080/log-error', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(logPayload),
        });

        if (!response.ok) {
            console.error('Failed to log client error:', await response.text());
        }
    } catch (err) {
        console.error('Network error while reporting client error:', err);
    }
}

window.addEventListener('error', (event) => {
    reportClientError(event.message, event.filename, event.lineno, event.colno, event.error?.stack || null);
});

window.addEventListener('unhandledrejection', (event) => {
    reportClientError(event.reason?.message || 'Unhandled rejection', null, null, null, event.reason?.stack || null);
});

