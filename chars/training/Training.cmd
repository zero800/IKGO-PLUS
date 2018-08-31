[Defaults]
command.time = 15
command.buffer.time = 1
;---�L�[�Q��A������------------------------------------

[Command]
name = "FF"       
command = F, F
time = 20

[Command]
name = "BB"       
command = B, B
time = 20

;---�󂯐g--------------------------------------------

[Command]
name = "recovery" 
command = y
time = 1

;---�����L�[�{�{�^��------------------------------------

[Command]
name = "down_a"
command = /$D,a
time = 1

[Command]
name = "down_b"
command = /$D,b
time = 1

[Command]
name = "down_c"
command = /$D,c
time = 1

;---�{�^���P��------------------------------------------

[Command]
name = "a"
command = a
time = 1

[Command]
name = "b"
command = b
time = 1

[Command]
name = "c"
command = c
time = 1

[Command]
name = "x"
command = x
time = 1

[Command]
name = "y"
command = y
time = 1

[Command]
name = "z"
command = z
time = 1

[Command]
name = "up"
command = U
time = 1

[Command]
name = "down"
command = D
time = 1

[Command]
name = "fwd"
command = F
time = 1

[Command]
name = "back"
command = B
time = 1

[Command]
name = "start"
command = s
time = 1

[Command]
name = "hold_x"
command = /x

[Command]
name = "hold_z"
command = /z


;---�����L�[ |---------------------------------------
[Command]
name = "holdfwd"   
command = /$F
time = 1

[Command]
name = "holdback"  
command = /$B
time = 1

[Command]
name = "holdup"    
command = /$U
time = 1

[Command]
name = "holddown"  
command = /$D
time = 1

[Statedef -1]
;�ړ�------------------------------------------------------------
[State -1, �_�b�V��]
type = ChangeState
triggerall = command = "FF"
trigger1 = (StateType != A) && (Ctrl)
value = 100

[State -1, �o�b�N�X�e�b�v]
type = ChangeState
triggerall = command = "BB"
trigger1 = (StateType != A) && (Ctrl)
value = 105

[State -1, �N���オ��]
type = ChangeState
triggerall = !var(35)
trigger1 = var(30) >= var(29)
trigger1 = stateno = 5110
value = 5120

[State -1, �ړ��N���オ��i�O�j]
type = ChangeState
triggerall = var(35) = 1
trigger1 = var(30) >= var(29)
trigger1 = stateno = 5110
value = 1000

[State -1, �ړ��N���オ��i���j]
type = ChangeState
triggerall = var(35) = 2
trigger1 = var(30) >= var(29)
trigger1 = stateno = 5110
value = 1001

;�I�[�g�K�[�h--------------------------------------------------------
[State -1,�K�[�h]
type = ChangeState
value = 120
triggerall = var(8)
triggerall = var(27) = 2
triggerall = StateNo!=[120,155]
triggerall = Ctrl||stateno = 21
trigger1 = inguarddist

[State -1,�K�[�h]
type = ChangeState
value = 120
triggerall = var(27) = 1 || var(27) = 3
triggerall = StateNo!=[120,155]
triggerall = Ctrl||stateno = 21
trigger1 = inguarddist

;---�ړ����̑�-----------------------------------------------------------------------------
[State -1, �A�h�K����]
type = ChangeState
value = 2000
triggerall = var(27) = 3
trigger1 = stateno = [150,151]

[State -1, �A�h�K��]
type = ChangeState
value = 2001
triggerall = var(27) = 3
trigger1 = stateno = [152,153]

[State -1, �A�h�K��]
type = ChangeState
value = 2002
triggerall = var(27) = 3
trigger1 = stateno = 154
trigger2 = stateno = 155 && time <= 10

;��{�s��------------------------------------------------------------
[State -1, ���̈ʒu�`�F�b�N�I]
type = ChangeState
triggerall = command = "a"
trigger1 = (StateType != A) && (Ctrl)
value = 200

[State -1, �̂̈ʒu�`�F�b�N�I]
type = ChangeState
triggerall = command = "b"
trigger1 = (StateType != A) && (Ctrl)
value = 201

[State -1, �˒������`�F�b�N�I]
type = ChangeState
triggerall = command = "c"
trigger1 = (StateType != A) && (Ctrl)
value = 300

[State -1, �J�E���g�_�E���A�^�b�N1]
type = ChangeState
triggerall = command = "x"
triggerall = command != "holdfwd" 
triggerall = command != "holdback" 
trigger1 = (StateType = S) && (Ctrl)
value = 400

[State -1, �J�E���g�_�E���A�^�b�N2]
type = ChangeState
triggerall = command = "x"
triggerall = command != "holdfwd" 
triggerall = command != "holdback" 
trigger1 = (StateType = C) && (Ctrl)
value = 405

[State -1, �J�E���g�_�E���A�^�b�N3]
type = ChangeState
triggerall = command = "x"
trigger1 = (StateType = A) && (Ctrl)
value = 406

[State -1, ��ѓ���P]
type = ChangeState
triggerall = command = "x"
triggerall = command = "holdfwd" 
trigger1 = (StateType != A) && (Ctrl)
value = 410

[State -1, ��ѓ���H]
type = ChangeState
triggerall = command = "x"
triggerall = command = "holdback" 
trigger1 = (StateType != A) && (Ctrl)
value = 420

[State -1, �p�����[�^�\���ؑ�]
type = varadd
trigger1 = command = "y"
var(17) = 1

[State -1, ���[���[�\���ؑ�]
type = varadd
trigger1 = command = "z"
var(21) = 1

;����s��----------------------------------------------------
[State -1, ����������]
type = ChangeState
triggerall = var(25)%4 != 0
triggerall = var(26)
trigger1 = (StateType != A) && (Ctrl)
value = 21

[State -1, ���Ⴊ��ő҂�]
type = ChangeState
triggerall = var(24)%3 = 1
triggerall = var(26) = 0
trigger1 = (StateType != A) && (Ctrl)
value = 11

[State -1, ����ő҂�]
type = ChangeState
triggerall = var(24)%3 = 2
trigger1 = (StateType != A) && (Ctrl)
trigger2 = stateno = 21
value = 40
