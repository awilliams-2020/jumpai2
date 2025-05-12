import React, { useState, useEffect } from 'react';
import {
	Box,
	Button,
	Card,
	CardContent,
	Container,
	Dialog,
	DialogActions,
	DialogContent,
	DialogTitle,
	Grid,
	TextField,
	Typography,
	Snackbar,
	Alert,
	IconButton,
	List,
	ListItem,
	ListItemText,
	CircularProgress,
	Collapse,
	Divider,
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import DeleteIcon from '@mui/icons-material/Delete';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import { DateTimePicker } from '@mui/x-date-pickers/DateTimePicker';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';
import { format } from 'date-fns';
import client from '../api/client';
import ConnectedAccounts from '../components/ConnectedAccounts';
import SchedulingWindows from '../components/SchedulingWindows';
import CalendarEvents from '../components/CalendarEvents';

interface Meeting {
	id: number;
	client_email: string;
	linkedin_url: string;
	start_time: string;
	end_time: string;
	answers: string[];
	context_notes: string;
}

interface SchedulingLink {
	id: string;
	title: string;
	duration: number;
	max_uses?: number;
	expires_at?: string;
	max_days_in_advance: number;
	custom_questions: string[];
	is_active: boolean;
	meetings?: Meeting[];
}

export default function Dashboard() {
	const [open, setOpen] = useState(false);
	const [links, setLinks] = useState<SchedulingLink[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [success, setSuccess] = useState<string | null>(null);
	const [formData, setFormData] = useState({
		title: '',
		duration: 30,
		maxUses: undefined as number | undefined,
		expiresAt: null as Date | null,
		maxDaysInAdvance: 30,
		customQuestions: [] as string[],
	});
	const [newQuestion, setNewQuestion] = useState('');
	const [expandedLinks, setExpandedLinks] = useState<{ [key: string]: boolean }>({});

	useEffect(() => {
		const fetchData = async () => {
			try {
				const linksResponse = await client.get('/api/scheduling/links');
				const linksWithMeetings = await Promise.all(
					linksResponse.data.map(async (link: SchedulingLink) => {
						try {
							const meetingsResponse = await client.get(`/api/scheduling/links/${link.id}/meetings`);
							return { ...link, meetings: meetingsResponse.data };
						} catch (err) {
							console.error(`Failed to fetch meetings for link ${link.id}:`, err);
							return { ...link, meetings: [] };
						}
					})
				);
				setLinks(linksWithMeetings);
			} catch (err: any) {
				console.error('Failed to fetch data:', err);
				setError(err.response?.data?.error || 'Failed to load data');
			} finally {
				setLoading(false);
			}
		};

		fetchData();
	}, []);

	const handleClickOpen = () => {
		setOpen(true);
	};

	const handleClose = () => {
		setOpen(false);
		setFormData({
			title: '',
			duration: 30,
			maxUses: undefined,
			expiresAt: null,
			maxDaysInAdvance: 30,
			customQuestions: [],
		});
		setNewQuestion('');
		setError(null);
	};

	const handleAddQuestion = () => {
		if (newQuestion.trim()) {
			setFormData({
				...formData,
				customQuestions: [...formData.customQuestions, newQuestion.trim()],
			});
			setNewQuestion('');
		}
	};

	const handleRemoveQuestion = (index: number) => {
		setFormData({
			...formData,
			customQuestions: formData.customQuestions.filter((_, i) => i !== index),
		});
	};

	const handleSubmit = async () => {
		try {
			const response = await client.post('/api/scheduling/links', {
				title: formData.title,
				duration: formData.duration,
				max_uses: formData.maxUses || null,
				expires_at: formData.expiresAt?.toISOString(),
				max_days_in_advance: formData.maxDaysInAdvance,
				custom_questions: formData.customQuestions,
			});

			setLinks([...links, response.data]);
			setSuccess('Scheduling link created successfully');
			handleClose();
		} catch (error) {
			console.error('Failed to create scheduling link:', error);
			setError('Failed to create scheduling link. Please try again.');
		}
	};

	const handleCopyLink = (id: string) => {
		const url = `${window.location.origin}/schedule/${id}`;
		navigator.clipboard.writeText(url);
		setSuccess('Link copied to clipboard');
	};

	const handleError = (message: string) => {
		setError(message);
	};

	const handleSuccess = (message: string) => {
		setSuccess(message);
	};

	const toggleLinkExpansion = (linkId: string) => {
		setExpandedLinks(prev => ({
			...prev,
			[linkId]: !prev[linkId]
		}));
	};

	if (loading) {
		return (
			<Box sx={{ display: 'flex', justifyContent: 'center', padding: '2rem' }}>
				<CircularProgress />
			</Box>
		);
	}

	return (
		<LocalizationProvider dateAdapter={AdapterDateFns}>
			<Container maxWidth="xl">
				<Box sx={{ mb: 4 }}>
					<ConnectedAccounts onError={handleError} onSuccess={handleSuccess} />
				</Box>

				<Box sx={{ mb: 4 }}>
					<SchedulingWindows />
				</Box>

				<Box sx={{ mb: 4 }}>
					<CalendarEvents />
				</Box>

				<Box sx={{ mb: 4 }}>
					<Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
						<Typography variant="h4" component="h1">
							Scheduling Links
						</Typography>
						<Button
							variant="contained"
							startIcon={<AddIcon />}
							onClick={handleClickOpen}
						>
							Create Link
						</Button>
					</Box>

					<Grid container spacing={2}>
						{links.map((link) => (
							<Grid item xs={12} md={6} lg={4} key={link.id}>
								<Card>
									<CardContent>
										<Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
											<Typography variant="h6" gutterBottom>
												{link.title}
											</Typography>
											<IconButton
												onClick={() => toggleLinkExpansion(link.id)}
												size="small"
											>
												{expandedLinks[link.id] ? <ExpandLessIcon /> : <ExpandMoreIcon />}
											</IconButton>
										</Box>
										<Typography variant="body2" color="text.secondary" gutterBottom>
											Duration: {link.duration} minutes
										</Typography>
										{link.max_uses && (
											<Typography variant="body2" color="text.secondary" gutterBottom>
												Max uses: {link.max_uses}
											</Typography>
										)}
										{link.expires_at && (
											<Typography variant="body2" color="text.secondary" gutterBottom>
												Expires: {format(new Date(link.expires_at), 'MMM d, yyyy')}
											</Typography>
										)}
										<Typography variant="body2" color="text.secondary" gutterBottom>
											Max days in advance: {link.max_days_in_advance}
										</Typography>
										<Box sx={{ mt: 2 }}>
											<Button
												variant="outlined"
												size="small"
												onClick={() => handleCopyLink(link.id)}
											>
												Copy Link
											</Button>
										</Box>

										<Collapse in={expandedLinks[link.id]}>
											<Box sx={{ mt: 2 }}>
												<Divider sx={{ my: 2 }} />
												<Typography variant="subtitle1" gutterBottom>
													Meetings ({link.meetings?.length || 0})
												</Typography>
												{link.meetings && link.meetings.length > 0 ? (
													<List>
														{link.meetings.map((meeting) => (
															<ListItem key={meeting.id} divider>
																<ListItemText
																	primary={
																		<Box component="span">
																			<Box component="span" sx={{ display: 'block', typography: 'subtitle2' }}>
																				{meeting.client_email}
																			</Box>
																			<Box component="span" sx={{ display: 'block', typography: 'body2', color: 'text.secondary' }}>
																				{format(new Date(meeting.start_time), 'MMM d, yyyy h:mm a')}
																			</Box>
																		</Box>
																	}
																	secondary={
																		<Box component="span" sx={{ mt: 1 }}>
																			{meeting.linkedin_url && (
																				<Box component="span" sx={{ display: 'block', typography: 'body2', color: 'text.secondary' }}>
																					LinkedIn: {meeting.linkedin_url}
																				</Box>
																			)}
																			{meeting.context_notes && (
																				Object.entries(JSON.parse(meeting.context_notes)).map(
																					([key, value]) => (
																						<Box key={key} component="span" sx={{ display: 'block', typography: 'body2', color: 'text.secondary' }}>
																							Notes: {String(value)}
																						</Box>
																					)
																				)
																			)}
																		</Box>
																	}
																/>
															</ListItem>
														))}
													</List>
												) : (
													<Typography variant="body2" color="text.secondary">
														No meetings scheduled yet
													</Typography>
												)}
											</Box>
										</Collapse>
									</CardContent>
								</Card>
							</Grid>
						))}
					</Grid>
				</Box>

				<Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
					<DialogTitle>Create Scheduling Link</DialogTitle>
					<DialogContent>
						<Box sx={{ pt: 2, display: 'flex', flexDirection: 'column', gap: 2 }}>
							<TextField
								label="Title"
								value={formData.title}
								onChange={(e) => setFormData({ ...formData, title: e.target.value })}
								fullWidth
								required
							/>

							<TextField
								label="Duration (minutes)"
								type="number"
								value={formData.duration}
								onChange={(e) => setFormData({ ...formData, duration: Number(e.target.value) })}
								inputProps={{ min: 15, step: 15 }}
								fullWidth
								required
							/>

							<TextField
								label="Max Uses"
								type="number"
								value={formData.maxUses || ''}
								onChange={(e) => setFormData({ ...formData, maxUses: Number(e.target.value) })}
								inputProps={{ min: 1 }}
								fullWidth
							/>

							<DateTimePicker
								label="Expires At"
								value={formData.expiresAt}
								onChange={(date) => setFormData({ ...formData, expiresAt: date })}
							/>

							<TextField
								label="Max Days in Advance"
								type="number"
								value={formData.maxDaysInAdvance}
								onChange={(e) => setFormData({ ...formData, maxDaysInAdvance: Number(e.target.value) })}
								inputProps={{ min: 1 }}
								fullWidth
								required
							/>

							<Box>
								<Typography variant="subtitle1" gutterBottom>
									Custom Questions
								</Typography>
								<Box sx={{ display: 'flex', gap: 1, mb: 1 }}>
									<TextField
										label="New Question"
										value={newQuestion}
										onChange={(e) => setNewQuestion(e.target.value)}
										fullWidth
									/>
									<Button
										variant="contained"
										onClick={handleAddQuestion}
										disabled={!newQuestion.trim()}
									>
										Add
									</Button>
								</Box>
								<List>
									{formData.customQuestions.map((question, index) => (
										<ListItem
											key={index}
											secondaryAction={
												<IconButton
													edge="end"
													aria-label="delete"
													onClick={() => handleRemoveQuestion(index)}
												>
													<DeleteIcon />
												</IconButton>
											}
										>
											<ListItemText primary={question} />
										</ListItem>
									))}
								</List>
							</Box>
						</Box>
					</DialogContent>
					<DialogActions>
						<Button onClick={handleClose}>Cancel</Button>
						<Button
							onClick={handleSubmit}
							variant="contained"
							disabled={!formData.title || formData.duration < 15}
						>
							Create
						</Button>
					</DialogActions>
				</Dialog>

				<Snackbar
					open={!!error}
					autoHideDuration={6000}
					onClose={() => setError(null)}
				>
					<Alert severity="error" onClose={() => setError(null)}>
						{error}
					</Alert>
				</Snackbar>

				<Snackbar
					open={!!success}
					autoHideDuration={6000}
					onClose={() => setSuccess(null)}
				>
					<Alert severity="success" onClose={() => setSuccess(null)}>
						{success}
					</Alert>
				</Snackbar>
			</Container>
		</LocalizationProvider>
	);
} 