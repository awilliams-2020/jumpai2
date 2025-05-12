import { Box, Button, Container, Grid, Typography } from '@mui/material';
import { useNavigate } from 'react-router-dom';
import CalendarMonthIcon from '@mui/icons-material/CalendarMonth';
import LinkIcon from '@mui/icons-material/Link';
import AnalyticsIcon from '@mui/icons-material/Analytics';

export default function Home() {
  const navigate = useNavigate();

  const features = [
    {
      icon: <CalendarMonthIcon sx={{ fontSize: 40 }} />,
      title: 'Smart Scheduling',
      description: 'Easily manage your availability and let clients book meetings at their convenience.',
    },
    {
      icon: <LinkIcon sx={{ fontSize: 40 }} />,
      title: 'Custom Booking Links',
      description: 'Create personalized booking links with custom forms and meeting durations.',
    },
    {
      icon: <AnalyticsIcon sx={{ fontSize: 40 }} />,
      title: 'CRM Integration',
      description: 'Seamlessly integrate with HubSpot and Google Calendar for better client management.',
    },
  ];

  return (
    <Box>
      {/* Hero Section */}
      <Box
        sx={{
          bgcolor: 'primary.main',
          color: 'white',
          py: 8,
          mb: 6,
        }}
      >
        <Container maxWidth="md">
          <Typography
            component="h1"
            variant="h2"
            align="center"
            gutterBottom
          >
            Schedule Smarter, Not Harder
          </Typography>
          <Typography variant="h5" align="center" paragraph>
            Streamline your client meetings with our intelligent scheduling platform.
            Connect your calendars, create custom booking links, and manage your
            client relationships all in one place.
          </Typography>
          <Box sx={{ mt: 4, textAlign: 'center' }}>
            <Button
              variant="contained"
              color="secondary"
              size="large"
              onClick={() => navigate('/login')}
            >
              Get Started
            </Button>
          </Box>
        </Container>
      </Box>

      {/* Features Section */}
      <Container maxWidth="lg">
        <Grid container spacing={4}>
          {features.map((feature, index) => (
            <Grid item xs={12} md={4} key={index}>
              <Box
                sx={{
                  textAlign: 'center',
                  p: 3,
                }}
              >
                <Box sx={{ color: 'primary.main', mb: 2 }}>
                  {feature.icon}
                </Box>
                <Typography variant="h5" component="h2" gutterBottom>
                  {feature.title}
                </Typography>
                <Typography color="text.secondary">
                  {feature.description}
                </Typography>
              </Box>
            </Grid>
          ))}
        </Grid>
      </Container>
    </Box>
  );
} 