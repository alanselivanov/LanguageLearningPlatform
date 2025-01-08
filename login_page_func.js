document.getElementById('loginForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    const name = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    try {
        const response = await fetch('/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ name, password })
        });
        
        if (response.ok) {
            const data = await response.json();
            
            localStorage.setItem('user', JSON.stringify({ id: data.id, name: data.name, role: data.role }));
            console.log(localStorage.getItem('user'))
            if (data.role === 'admin') {
                window.location.href = '/static/index.html';
            } else {
                window.location.href = '/main_page.html';
            }
        } else {
            alert('Login failed');
        }
    } catch (error) {
        console.error('Error during login:', error);
        alert('An error occurred. Please try again later.');
    }
});
