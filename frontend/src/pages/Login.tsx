import { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Box, Button, Container, Typography, Alert } from '@mui/material';
import GoogleIcon from '@mui/icons-material/Google';
import client from '../api/client';
import type { User } from '../types/user';

export default function Login() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const token = searchParams.get('token');
    if (token) {
      // Store token without Bearer prefix
      const rawToken = token.startsWith('Bearer ') ? token.slice(7) : token;
      localStorage.setItem('token', rawToken);
      // Set token in client headers with Bearer prefix
      client.defaults.headers.common['Authorization'] = `Bearer ${rawToken}`;
      
      // Fetch user profile
      const fetchProfile = async () => {
        try {
          const response = await client.get<User>('/api/google/profile');
          if (response.data) {
            // Store profile data
            localStorage.setItem('userProfile', JSON.stringify(response.data));
            // Navigate to dashboard
            navigate('/dashboard');
          }
        } catch (error: any) {
          console.error('Failed to fetch profile:', error);
          setError(error.response?.data?.error || 'Failed to fetch profile');
          // Clear invalid token
          localStorage.removeItem('token');
          localStorage.removeItem('userProfile');
        }
      };
      
      fetchProfile();
    }
  }, [searchParams, navigate]);

  const handleGoogleLogin = () => {
    window.location.href = '/auth/google/login';
  };

  return (
    <Container maxWidth="sm">
      <Box
        sx={{
          marginTop: 8,
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
        }}
      >
        <Typography component="h1" variant="h4" gutterBottom>
          Welcome Back
        </Typography>
        <Typography variant="body1" color="text.secondary" align="center" paragraph>
          Sign in to manage your scheduling links and client meetings
        </Typography>
        
        {error && (
          <Alert severity="error" sx={{ width: '100%', mb: 2 }}>
            {error}
          </Alert>
        )}

        <Button
          variant="contained"
          startIcon={<GoogleIcon />}
          onClick={handleGoogleLogin}
          sx={{ mt: 3, mb: 2 }}
          fullWidth
        >
          Continue with Google
        </Button>
      </Box>
    </Container>
  );
} 