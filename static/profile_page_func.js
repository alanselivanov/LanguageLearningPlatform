document.addEventListener('DOMContentLoaded', async () => {
    const token = localStorage.getItem('token');
    const user = JSON.parse(localStorage.getItem('user'));

    if (!token || !user) {
        alert('You need to log in first.');
        window.location.href = '/';
        return;
    }

    const form = document.getElementById('profileForm');
    const usernameField = document.getElementById('username');
    const emailField = document.getElementById('email');
    const passwordField = document.getElementById('password');

    try {
        const response = await fetch(`/readByIDprof?id=${user.id}`, {
            headers: {
                'Authorization': `Bearer ${token}`,
            },
        });
        if (response.ok) {
            const data = await response.json();
            usernameField.value = data.name;
            emailField.value = data.email;
            passwordField.value = data.password;
        } else {
            if (response.status === 401) {
                alert('Session expired. Please log in again.');
                localStorage.removeItem('token');
                localStorage.removeItem('user');
                window.location.href = '/';
            } else {
                alert('Failed to load profile data.');
            }
        }
    } catch (error) {
        console.error('Error fetching profile data:', error);
    }

    form.addEventListener('submit', async (e) => {
        e.preventDefault();
        const updatedData = {
            id: user.id,
            name: usernameField.value,
            email: emailField.value,
            password: passwordField.value,
        };

        try {
            const response = await fetch('/update', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`,
                },
                body: JSON.stringify(updatedData),
            });

            if (response.ok) {
                alert('Profile updated successfully!');
            } else {
                alert('Failed to update profile.');
            }
        } catch (error) {
            console.error('Error updating profile:', error);
        }
    });

    const logoutButton = document.getElementById('logoutButton');
    logoutButton.addEventListener('click', () => {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        window.location.href = '/';
    });
});
