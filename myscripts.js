async function createUser() {
    try {
        const name = document.getElementById('name')?.value.trim();
        const email = document.getElementById('email')?.value.trim();
        const password = document.getElementById('password')?.value.trim();
        const role = document.getElementById('role')?.value.trim().toLowerCase();

        if (!name) throw new Error('Name is required.');
        if (!email || !/^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$/i.test(email)) throw new Error('Invalid email format.');
        if (!password || password.length < 6) throw new Error('Password must be at least 6 characters long.');
        if (!role || (role !== "user" && role !== "admin")) throw new Error('Should be admin or user.');

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


async function getUsers() {

    try {
        const response = await fetch('/read');
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
        document.getElementById('output').innerHTML = output;
    } catch (err) {
        console.error('Error in getUsers:', err);
        await reportClientError(err.message, 'getUsers', null, null, err.stack || null);
        alert(`Failed to fetch users: ${err.message}`);
    }

    const response = await fetch('/read');
    const users = await response.json();
    let output = '<table border="1"><tr><th>ID</th><th>Name</th><th>Email</th><th>Password</th><th>Role</th><th>Created At</th><th>Updated At</th></tr>';
    users.forEach(user => {
        output += `<tr>
            <td>${user.id}</td>
            <td>${user.name}</td>
            <td>${user.email}</td>
            <td>${user.password}</td>
            <td>${user.role}</td>
            <td>${user.created_at}</td>
            <td>${user.updated_at}</td>
        </tr>`;
    });
    output += '</table>';
    document.getElementById('output').innerHTML = output;

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

