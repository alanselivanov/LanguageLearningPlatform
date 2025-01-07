async function createUser() {
    const name = document.getElementById('name').value.trim();
    const email = document.getElementById('email').value.trim();
    const password = document.getElementById('password').value.trim();
    const role = document.getElementById('role').value.trim().toLowerCase();

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
    if (!role || (role !== "user" && role !== "admin")) {
        alert('Role is required and must be either "user" or "admin".');
        return;
    }
    const response = await fetch('/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, email, password, role }),
    });

    if (response.ok) {
        const result = await response.json();
        alert(`User created successfully: ${JSON.stringify(result)}`);
    } else {
        const error = await response.text();
        alert(`Error creating user: ${error}`);
    }
}

async function getUsers() {
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
    const id = prompt('Enter User ID to update:');
    if (!id || isNaN(id) || parseInt(id) <= 0) {
        alert('Invalid User ID. Please enter a positive number.');
        return;
    }

    const name = prompt('Enter new name (leave blank to keep current):');
    if (name && name.length < 3) {
        alert('Name must be at least 3 characters long.');
        return;
    }

    const email = prompt('Enter new email (leave blank to keep current):');
    if (email && !/^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$/i.test(email)) {
        alert('Invalid email format.');
        return;
    }

    const password = prompt('Enter new password (leave blank to keep current):');
    if (password && password.length < 6) {
        alert('Password must be at least 6 characters long.');
        return;
    }

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

    if (response.ok) {
        const result = await response.json();
        alert(`User updated successfully: ${JSON.stringify(result)}`);
    } else {
        const error = await response.text();
        alert(`Error updating user: ${error}`);
    }
}

async function deleteUser() {
    const id = prompt('Enter User ID to delete:');
    if (!id || isNaN(id) || parseInt(id) <= 0) {
        alert('Invalid User ID. Please enter a positive number.');
        return;
    }

    const response = await fetch('/delete', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: parseInt(id) }),
    });

    if (response.ok) {
        const result = await response.json();
        alert(`User deleted successfully: ${JSON.stringify(result)}`);
    } else {
        const error = await response.text();
        alert(`Error deleting user: ${error}`);
    }
}

async function getUserByID() {
    const id = document.getElementById('userID').value;
    if (!id) {
        alert('Please enter a User ID');
        return;
    }
    const response = await fetch(`/readByID?id=${id}`);
    if (response.ok) {
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
    } else {
        const error = await response.text();
        alert(`Error: ${error}`);
    }
}
