
void FUN_0806fe4a(RedNum *param_1)

{
  Point *in_EAX;
  undefined1 *puVar1;
  int in_ECX;
  Point *in_EDX;
  undefined1 local_58 [60];
  undefined1 local_1c [12];
  
  puVar1 = local_58;
  do {
    FUN_0806f33c();
    *(undefined8 **)(puVar1 + 8) = &RedNum::emptyReducer;
    puVar1 = puVar1 + 0xc;
  } while (puVar1 != local_1c);
  FUN_0805d150();
  *(undefined4 *)(in_EAX + 8) = *(undefined4 *)(in_ECX + 8);
  FUN_0805d150();
  *(undefined4 *)(in_EAX + 0x14) = *(undefined4 *)(in_ECX + 0x14);
  FUN_0805d150();
  *(undefined4 *)(in_EAX + 0x20) = *(undefined4 *)(in_ECX + 0x20);
  puVar1 = local_1c;
  WCurve::add(in_EDX,in_EAX,param_1);
  do {
    puVar1 = puVar1 + -0xc;
    FUN_08052f66();
  } while (puVar1 != local_58);
  return;
}


