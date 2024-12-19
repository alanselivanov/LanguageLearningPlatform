async function createUser() {
    const name = document.getElementById('name').value;
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;

    // Получаем список пользователей, чтобы определить следующий ID
    const response = await fetch('/read');
    const users = await response.json();

    // Сортируем пользователей по ID и находим следующий ID
    const nextId = users.length ? Math.max(...users.map(user => user.id)) + 1 : 1;

    const createResponse = await fetch('/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: nextId, name, email, password }),
    });
    
    const result = await createResponse.json();
    alert(JSON.stringify(result));
}

async function getUsers() {
    const response = await fetch('/read');
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
}

async function updateUser() {
    const id = prompt('Enter User ID to update:');
    const name = prompt('Enter new name:');
    const email = prompt('Enter new email:');
    const password = prompt('Enter new password:');

    const response = await fetch('/read');
    const users = await response.json();
    const userExists = users.some(user => user.id === parseInt(id));

    if (!userExists) {
        alert('User ID not found!');
        return;
    }

    const updateResponse = await fetch('/update', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: parseInt(id), name, email, password }),
    });

    if (updateResponse.ok) {
        const result = await updateResponse.json();
        alert(`User updated successfully: ${JSON.stringify(result)}`);
    } else {
        const error = await updateResponse.text();
        alert(`Error updating user: ${error}`);
    }
}

async function deleteUser() {
    const id = prompt('Enter User ID to delete:');
    
    const response = await fetch('/read');
    const users = await response.json();
    const userExists = users.some(user => user.id === parseInt(id));

    if (!userExists) {
        alert('User ID not found!');
        return;
    }

    const deleteResponse = await fetch('/delete', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: parseInt(id) }),
    });

    const result = await deleteResponse.json();
    alert(JSON.stringify(result));
}
