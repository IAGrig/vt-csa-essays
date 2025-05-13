import React, { useState } from 'react';
import axios from 'axios';
import { Box, TextField, Button, Typography, Alert } from '@mui/material';
import { setAccessToken } from '../auth';
import { useNavigate } from 'react-router-dom';
import { API_BASE_URL } from '../config';

const ALPHANUM = /^[A-Za-z0-9]+$/;

const Register: React.FC = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [formError, setFormError] = useState<string|null>(null);
  const navigate = useNavigate();

  const validate = () => {
    if (!username || !ALPHANUM.test(username)) {
      return 'Username must be alphanumeric and non-empty';
    }
    if (!password || !ALPHANUM.test(password)) {
      return 'Password must be alphanumeric and non-empty';
    }
    return null;
  };

  const submit = async () => {
    const err = validate();
    if (err) {
      setFormError(err);
      return;
    }
    setFormError(null);

    try {
      const { data } = await axios.post(
        `${API_BASE_URL}/user`,
        { username, password },
        { withCredentials: true }
      );
      setAccessToken(data.access_token);
      localStorage.setItem('username', username);
      navigate('/');
    } catch (e: any) {
      setFormError(e.response?.data?.error || 'Registration failed');
    }
  };

  return (
    <Box maxWidth={360} mx="auto">
      <Typography variant="h5" gutterBottom>Register</Typography>
      {formError && <Alert severity="error" sx={{ mb: 2 }}>{formError}</Alert>}
      <TextField
        fullWidth label="Username"
        value={username}
        onChange={e => setUsername(e.target.value)}
        sx={{ mb: 2 }}
      />
      <TextField
        fullWidth type="password" label="Password"
        value={password}
        onChange={e => setPassword(e.target.value)}
        sx={{ mb: 2 }}
      />
      <Button variant="contained" fullWidth onClick={submit}>
        Register
      </Button>
    </Box>
  );
};

export default Register;
