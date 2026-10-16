
void FUN_0805efca(void)

{
  char cVar1;
  undefined1 uVar2;
  undefined4 *in_EAX;
  BigNum *in_ECX;
  int in_EDX;
  undefined1 *puVar3;
  undefined1 auStack_8c [8];
  BigNum *local_84;
  int local_80;
  undefined1 local_70 [8];
  Point local_68 [88];
  
  puVar3 = auStack_8c;
  local_80 = in_EDX;
  FUN_08050d86();
  cVar1 = FUN_0804fe56();
  uVar2 = 0;
  if (cVar1 != '\0') {
    uVar2 = FUN_0804fe56();
  }
  local_84 = (BigNum *)CONCAT31(local_84._1_3_,uVar2);
  FUN_0804f6cc();
  if ((char)local_84 == '\0') {
    local_84 = (BigNum *)(local_80 + 0x30);
    cVar1 = BigNum::operator<(in_ECX,local_84);
    if ((cVar1 != '\0') && (cVar1 = BigNum::operator<(in_ECX + 8,local_84), cVar1 != '\0')) {
      FUN_0805b76a();
      cVar1 = FUN_0805ef2a();
      if (cVar1 == '\0') {
        *in_EAX = 0;
        in_EAX[1] = 0;
      }
      else {
        FUN_08059380();
        WCurve::toEC(local_68);
        BigNum::bin((uint)local_70);
        BigNum::bin((uint)in_EAX);
        puVar3 = &stack0xffffff68;
        FUN_0805091c();
        FUN_0804f6cc();
        FUN_08059352();
      }
      *(undefined4 *)(puVar3 + -4) = 0x805f0d6;
      FUN_0805931c();
      return;
    }
  }
  *in_EAX = 0;
  in_EAX[1] = 0;
  return;
}


