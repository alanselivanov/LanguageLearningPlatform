document.getElementById('supportForm').addEventListener('submit', async function (e) {
    e.preventDefault();

    const user = JSON.parse(localStorage.getItem('user')); 
    const token = localStorage.getItem('token');
    //if (!user) { 
        //alert('Error: You must be logged in to submit a file!');
        //return; 
    //}

    const formData = new FormData(this);

    try {
        const response = await fetch('/send-support-ticket', {
            method: 'POST',
            headers:{
                'Authorization': `Bearer ${token}`,
            },
            body: formData
        });

        if (response.ok) {
            alert('Message sent successfully!');
        } else {
            alert('Failed to send message');
        }
    } catch (error) {
        console.error('Error sending message:', error);
        alert('Error sending message');
    }
});

window.onload = function() {
    const user = JSON.parse(localStorage.getItem('user'));
    
    if (user) {
        document.getElementById('login-btn').innerText = 'Logout';
        document.getElementById('profile-btn').innerText = 'Profile';

        document.getElementById('login-btn').onclick = function(e) {
            e.preventDefault(); 
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            window.location.href = '/';
        };

        document.getElementById('profile-btn').onclick = function(e) {
            e.preventDefault();
            window.location.href = '/profilePage';
        };
    } else {
        document.getElementById('login-btn').innerText = 'Login';
        document.getElementById('profile-btn').innerText = 'Sign-up';

        document.getElementById('login-btn').onclick = function() {
            window.location.href = '/static/loginPage';
        };

        document.getElementById('profile-btn').onclick = function() {
            window.location.href = '/static/signupPage';
        };
    }
};
