async function createUser() {
    const name = document.getElementById('name').value;
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const response = await fetch('/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, email, password }),
    });
    const result = await response.json();
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
    const response = await fetch('/update', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: parseInt(id), name, email, password }),
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
    const response = await fetch('/delete', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: parseInt(id) }),
    });
    const result = await response.json();
    alert(JSON.stringify(result));
}