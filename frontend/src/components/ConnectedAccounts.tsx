import { useState, useEffect } from 'react';
import {
	Box,
	Button,
	Typography,
	Paper,
	List,
	ListItem,
	ListItemText,
	IconButton,
	Alert,
	Avatar,
} from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import HubIcon from '@mui/icons-material/Hub';
import GoogleIcon from '@mui/icons-material/Google';
import client from '../api/client';

interface HubSpotAccount {
	id: string;
	hub_name: string;
	hub_domain: string;
	hub_timezone: string;
	last_synced_at: string;
}

interface GoogleAccount {
	id: string;
	google_id: string;
	email: string;
	name: string;
	profile_picture: string;
	last_synced_at: string;
	is_active: boolean;
	calendar_ids: string[];
}

interface ConnectedAccountsProps {
	onError: (message: string) => void;
	onSuccess: (message: string) => void;
}

export default function ConnectedAccounts({ onError, onSuccess }: ConnectedAccountsProps) {
	const [hubspotAccounts, setHubspotAccounts] = useState<HubSpotAccount[]>([]);
	const [googleAccounts, setGoogleAccounts] = useState<GoogleAccount[]>([]);

	useEffect(() => {
		fetchConnectedAccounts();
	}, []);

	const fetchConnectedAccounts = async () => {
		try {
			const [hubspotResponse, googleResponse] = await Promise.all([
				client.get('/api/hubspot/accounts'),
				client.get('/api/google/accounts')
			]);
			setHubspotAccounts(hubspotResponse.data);
			setGoogleAccounts(googleResponse.data);
		} catch (error) {
			console.error('Failed to fetch connected accounts:', error);
		}
	};

	const handleConnectHubSpot = () => {
		const token = localStorage.getItem('token');
		if (!token) {
			onError('Authentication token not found');
			return;
		}
		window.location.href = `/api/hubspot/connect?token=${token}`;
	};

	const handleConnectGoogle = () => {
		const token = localStorage.getItem('token');
		if (!token) {
			onError('Authentication token not found');
			return;
		}
		window.location.href = `/api/google/connect?token=${token}`;
	};

	const handleDisconnectHubSpot = async (accountId: string) => {
		try {
			await client.delete(`/api/hubspot/accounts/${accountId}`);
			setHubspotAccounts(accounts => accounts.filter(acc => acc.id !== accountId));
			onSuccess('HubSpot account disconnected successfully');
		} catch (error) {
			console.error('Failed to disconnect HubSpot account:', error);
			onError('Failed to disconnect HubSpot account');
		}
	};

	const handleDisconnectGoogle = async (accountId: string) => {
		try {
			await client.delete(`/api/google/accounts/${accountId}`);
			setGoogleAccounts(accounts => accounts.filter(acc => acc.id !== accountId));
			onSuccess('Google account disconnected successfully');
		} catch (error) {
			console.error('Failed to disconnect Google account:', error);
			onError('Failed to disconnect Google account');
		}
	};

	return (
		<Box sx={{ mb: 4 }}>
			<Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
				<Typography variant="h4" component="h1">
					Connected Accounts
				</Typography>
				<Box sx={{ display: 'flex', gap: 2 }}>
					<Button
						variant="contained"
						startIcon={<GoogleIcon />}
						onClick={handleConnectGoogle}
					>
						Connect Google
					</Button>
					<Button
						variant="contained"
						startIcon={<HubIcon />}
						onClick={handleConnectHubSpot}
					>
						Connect HubSpot
					</Button>
				</Box>
			</Box>
			<Paper sx={{ p: 2 }}>
				{hubspotAccounts.length === 0 && googleAccounts.length === 0 ? (
					<Typography color="text.secondary" align="center" sx={{ py: 2 }}>
						No accounts connected. Connect a Google or HubSpot account to get started.
					</Typography>
				) : (
					<>
						{googleAccounts.length > 0 && (
							<>
								<Typography variant="h6" sx={{ mt: 2, mb: 1 }}>
									Google Accounts
								</Typography>
								<List>
									{googleAccounts.map((account) => (
										<ListItem
											key={account.id}
											secondaryAction={
												<IconButton
													edge="end"
													aria-label="disconnect"
													onClick={() => handleDisconnectGoogle(account.id)}
												>
													<DeleteIcon />
												</IconButton>
											}
										>
											<Avatar
												src={account.profile_picture}
												sx={{ mr: 2 }}
											>
												<GoogleIcon />
											</Avatar>
											<ListItemText
												primary={account.name}
												secondary={account.email}
											/>
										</ListItem>
									))}
								</List>
							</>
						)}

						{hubspotAccounts.length > 0 && (
							<>
								<Typography variant="h6" sx={{ mt: 2, mb: 1 }}>
									HubSpot Accounts
								</Typography>
								<List>
									{hubspotAccounts.map((account) => (
										<ListItem
											key={account.id}
											secondaryAction={
												<IconButton
													edge="end"
													aria-label="disconnect"
													onClick={() => handleDisconnectHubSpot(account.id)}
												>
													<DeleteIcon />
												</IconButton>
											}
										>
											<Avatar sx={{ mr: 2, bgcolor: 'primary.main' }}>
												<HubIcon />
											</Avatar>
											<ListItemText
												primary={account.hub_name}
												secondary={`Domain: ${account.hub_domain} • Timezone: ${account.hub_timezone}`}
											/>
										</ListItem>
									))}
								</List>
							</>
						)}
					</>
				)}
			</Paper>
		</Box>
	);
} 